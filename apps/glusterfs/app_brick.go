package glusterfs

import (
"encoding/json"
"fmt"
"math"
"net/http"

"github.com/boltdb/bolt"
"github.com/gorilla/mux"
"github.com/heketi/heketi/pkg/db"
"github.com/heketi/heketi/pkg/glusterfs/api"
"github.com/heketi/heketi/pkg/utils"
)

const (
	VOLUME_CREATE_MAX_SNAPSHOT_FACTOR = 100
)


func (a *App) BrickDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var brick *BrickEntry
	err := a.db.View(func(tx *bolt.Tx) error {

		var err error
		brick, err = NewBrickEntryFromId(tx, id)
		if err == ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return err
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		return nil

	})
	if err != nil {
		return
	}

	a.asyncManager.AsyncHttpRedirectFunc(w, r, func() (string, error) {

		// Actually destroy the Brick here
		err := brick.Destroy(a.db, a.executor)

		// If it fails for some reason, we will need to add to the DB again
		// or hold state on the entry "DELETING"

		// Show that the key has been deleted
		if err != nil {
			logger.LogError("Failed to delete brick %v: %v", brick.Info.Id, err)
			return "", err
		}

		logger.Info("Deleted brick [%s]", id)
		return "", nil

	})

}

