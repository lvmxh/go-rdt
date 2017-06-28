package db

import (
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"openstackcore-rdtagent/model/workload"
)

// mgo database session
var session *mgo.Session

// Global variable for Database name
var Dbname string

type MgoDB struct {
}

func (m *MgoDB) Initialize(transport, dbname string) error {

	Dbname = dbname

	session, err := mgo.Dial(transport)

	if err != nil {
		return err
	}

	c := session.DB(Dbname).C(WorkloadTableName)
	if c == nil {
		return errors.New("Unable to create collection RDTpolicy")
	}

	index := mgo.Index{
		Key:    []string{"ID"},
		Unique: true,
	}

	err = c.EnsureIndex(index)

	if err != nil {
		return err
	}
	return nil
}

func (m *MgoDB) CreateWorkload(w *workload.RDTWorkLoad) error {
	s := session.Copy()
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

func (m *MgoDB) GetAllWorkload() ([]workload.RDTWorkLoad, error) {
	ws := []workload.RDTWorkLoad{}
	s := session.Copy()
	defer s.Close()

	if err := s.DB(Dbname).C(WorkloadTableName).Find(nil).All(&ws); err != nil {
		return ws, err
	}

	return ws, nil
}

func (m *MgoDB) GetWorkloadById(id string) (workload.RDTWorkLoad, error) {
	w := workload.RDTWorkLoad{}
	s := session.Copy()
	defer s.Close()

	if err := s.DB(Dbname).C(WorkloadTableName).Find(bson.M{"ID": w.ID}).One(&w); err != nil {
		return w, err
	}

	return w, nil

}
