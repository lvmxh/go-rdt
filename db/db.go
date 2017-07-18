package db

import (
	"fmt"
	"reflect"

	"openstackcore-rdtagent/lib/util"
	"openstackcore-rdtagent/model/workload"
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

// Helper function to find if a elem in a slice
func hasElem(s interface{}, elem interface{}) bool {
	arrv := reflect.ValueOf(s)
	if arrv.Kind() == reflect.Slice {
		for i := 0; i < arrv.Len(); i++ {
			if arrv.Index(i).Interface() == elem {
				return true
			}
		}
	}
	return false
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
			if hasElem(wi.TaskIDs, t) {
				return fmt.Errorf("Taskid %s has existed in workload %s", t, wi.ID)
			}
		}
	}

	if len(w.CoreIDs) == 0 {
		return nil
	}

	// Validate if the core id of workload has overlap with crrent ones.
	bm, _ := util.NewBitmap(w.CoreIDs)
	bmsum, _ := util.NewBitmap("")

	for _, c := range ws {
		if len(c.CoreIDs) > 0 {
			tmpbm, _ := util.NewBitmap(c.CoreIDs)
			bmsum = bmsum.Or(tmpbm)
		}
	}

	bminter := bm.And(bmsum)

	if !bminter.IsEmpty() {
		return fmt.Errorf("CPU list %s has been assigned.", bminter.ToString())
	}

	return nil
}
