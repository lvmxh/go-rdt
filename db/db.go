package db

import (
	"fmt"

	// from app import an config is really not a good idea.
	// uncouple it from APP. Or we can add it in a rmd/config
	. "openstackcore-rdtagent/db/config"
	libutil "openstackcore-rdtagent/lib/util"
	"openstackcore-rdtagent/model/workload"
	"openstackcore-rdtagent/util"
)

// workload table name
const WorkloadTableName = "workload"

type DB interface {
	Initialize(transport, dbname string) error
	CreateWorkload(w *workload.RDTWorkLoad) error
	DeleteWorkload(w *workload.RDTWorkLoad) error
	UpdateWorkload(w *workload.RDTWorkLoad) error
	GetAllWorkload() ([]workload.RDTWorkLoad, error)
	GetWorkloadById(id string) (workload.RDTWorkLoad, error)
	ValidateWorkload(w *workload.RDTWorkLoad) error
}

func NewDB() (DB, error) {
	dbcon := NewConfig()
	if dbcon.Backend == "bolt" {
		return newBoltDB()
	} else if dbcon.Backend == "mgo" {
		return newMgoDB()
	} else {
		return nil, fmt.Errorf("Unsupported DB backend %s", dbcon.Backend)
	}
}

// this function does 3 things to valicate a user request workload is
// validate at data base layer
func validateWorkload(w workload.RDTWorkLoad, ws []workload.RDTWorkLoad) error {

	if len(w.ID) < 1 && len(w.TaskIDs) < 1 && len(w.CoreIDs) < 1 {
		return nil
	}

	// User post a workload id/uuid in it's request
	if w.ID != "" {
		for _, i := range ws {
			if w.ID == i.ID {
				return fmt.Errorf("Workload id %s has existed %s", w.ID)
			}
		}
	}

	// Validate if the task id of workload has existed.
	for _, t := range w.TaskIDs {
		for _, wi := range ws {
			if util.HasElem(wi.TaskIDs, t) {
				return fmt.Errorf("Taskid %s has existed in workload %s", t, wi.ID)
			}
		}
	}

	if len(w.CoreIDs) == 0 {
		return nil
	}

	// Validate if the core id of workload has overlap with crrent ones.
	bm, _ := libutil.NewBitmap(w.CoreIDs)
	bmsum, _ := libutil.NewBitmap("")

	for _, c := range ws {
		if len(c.CoreIDs) > 0 {
			tmpbm, _ := libutil.NewBitmap(c.CoreIDs)
			bmsum = bmsum.Or(tmpbm)
		}
	}

	bminter := bm.And(bmsum)

	if !bminter.IsEmpty() {
		return fmt.Errorf("CPU list %s has been assigned.", bminter.ToString())
	}

	return nil
}
