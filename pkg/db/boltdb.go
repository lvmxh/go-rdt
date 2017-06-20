package db

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"log"
	"strconv"
	"sync"

	"openstackcore-rdtagent/pkg/model/workload"
)

var once sync.Once
var lock sync.Mutex
var db *bolt.DB

type BoltDB struct {
}

const WorkloadBucket = "workload"

func initDB(dbname string) error {
	var err error
	db, err = bolt.Open(dbname, 0600, nil)
	if err != nil {
		return err
	}

	db.Update(func(tx *bolt.Tx) error {
		// First touch a Bucket
		_, err := tx.CreateBucketIfNotExists([]byte(WorkloadBucket))
		if err != nil {
			return err
		}
		return nil
	})

	return nil
}

func (b *BoltDB) Initialize(dbname string) {
	once.Do(func() {
		err := initDB(dbname)
		if err != nil {
			log.Panic(err)
		}
	})
}

func (b *BoltDB) CreateWorkload(w *workload.RDTWorkLoad) error {
	/* When create a new workload we need to verify that the new PIDs
	   we the workload specified should not existed */

	ws, err := b.GetAllWorkload()
	if err != nil {
		return err
	}

	err = validateTasks(*w, ws)
	if err != nil {
		return err
	}

	lock.Lock()
	defer lock.Unlock()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorkloadBucket))

		// Generate ID for the workload.
		id, _ := b.NextSequence()
		w.ID = strconv.Itoa(int(id))

		// Marshal  data into bytes.
		buf, err := json.Marshal(w)
		if err != nil {
			return err
		}
		// Persist bytes to users bucket.
		return b.Put([]byte(w.ID), buf)
	})
}

func (b *BoltDB) DeleteWorkload(w *workload.RDTWorkLoad) error {
	lock.Lock()
	defer lock.Unlock()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorkloadBucket))
		return b.Delete([]byte(w.ID))
	})
}

func (b *BoltDB) GetAllWorkload() ([]workload.RDTWorkLoad, error) {
	ws := []workload.RDTWorkLoad{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorkloadBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			w := workload.RDTWorkLoad{}
			json.Unmarshal(v, &w)
			ws = append(ws, w)
		}
		return nil
	})
	return ws, err
}

func (b *BoltDB) GetWorkloadById(id string) (workload.RDTWorkLoad, error) {
	w := workload.RDTWorkLoad{}
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WorkloadBucket))
		v := b.Get([]byte(id))
		return json.Unmarshal(v, &w)
	})
	return w, err
}
