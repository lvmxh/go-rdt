package db

import (
	"errors"
	"sync"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	. "openstackcore-rdtagent/db/config"
	"openstackcore-rdtagent/model/types/workload"
)

// mgo database session
var mgoSession *mgo.Session
var mgoSessionOnce sync.Once

// Global variable for Database name
var Dbname string

type MgoDB struct {
	session *mgo.Session
}

// We thought, open a file, means open a session.
// Unity Concept with mongodb
func getMgoSession() error {
	var err error
	mgoSessionOnce.Do(func() {
		conf := NewConfig()
		mgoSession, err = mgo.Dial(conf.Transport)
	})
	return err
}

func closeMgoSession() {
}

func newMgoDB() (DB, error) {
	var db MgoDB
	if err := getSession(); err != nil {
		return &db, err
	}
	db.session = mgoSession
	return &db, nil

}

func (m *MgoDB) Initialize(transport, dbname string) error {

	conf := NewConfig()
	// FIXME, Dbname here seems some urgly
	Dbname = conf.DBName

	c := m.session.DB(Dbname).C(WorkloadTableName)
	if c == nil {
		return errors.New("Unable to create collection RDTpolicy")
	}

	index := mgo.Index{
		Key:    []string{"ID"},
		Unique: true,
	}

	err := c.EnsureIndex(index)

	if err != nil {
		return err
	}
	return nil
}

func (m *MgoDB) ValidateWorkload(w *workload.RDTWorkLoad) error {
	/* When create a new workload we need to verify that the new PIDs
	   we the workload specified should not existed */
	// not implement yet
	return nil
}

func (m *MgoDB) CreateWorkload(w *workload.RDTWorkLoad) error {
	s := m.session.Copy()
	defer s.Close()

	if err := s.DB(Dbname).C(WorkloadTableName).Insert(w); err != nil {
		return err
	}
	return nil
}

func (m *MgoDB) DeleteWorkload(w *workload.RDTWorkLoad) error {
	// not implement yet
	return nil
}

func (m *MgoDB) UpdateWorkload(w *workload.RDTWorkLoad) error {
	// not implement yet
	return nil
}

func (m *MgoDB) GetAllWorkload() ([]workload.RDTWorkLoad, error) {
	ws := []workload.RDTWorkLoad{}
	s := m.session.Copy()
	defer s.Close()

	if err := s.DB(Dbname).C(WorkloadTableName).Find(nil).All(&ws); err != nil {
		return ws, err
	}

	return ws, nil
}

func (m *MgoDB) GetWorkloadById(id string) (workload.RDTWorkLoad, error) {
	w := workload.RDTWorkLoad{}
	s := m.session.Copy()
	defer s.Close()

	if err := s.DB(Dbname).C(WorkloadTableName).Find(bson.M{"ID": w.ID}).One(&w); err != nil {
		return w, err
	}

	return w, nil

}

func (m *MgoDB) QueryWorkload(query map[string]interface{}) ([]workload.RDTWorkLoad, error) {
	// not implement yet
	return []workload.RDTWorkLoad{}, nil
}
