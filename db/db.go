package db

import (
	"fmt"
	"reflect"

	"openstackcore-rdtagent/model/workload"
)

// workload table name
const WorkloadTableName = "workload"

type DB interface {
	Initialize(transport, dbname string) error
	CreateWorkload(w *workload.RDTWorkLoad) error
	DeleteWorkload(w *workload.RDTWorkLoad) error
	GetAllWorkload() ([]workload.RDTWorkLoad, error)
	GetWorkloadById(id string) (workload.RDTWorkLoad, error)
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

// Check if TasksIDs of a workload existed in a workload array
func validateTasks(w workload.RDTWorkLoad, ws []workload.RDTWorkLoad) error {
	if len(w.TaskIDs) < 1 {
		return nil
	}

	for _, t := range w.TaskIDs {
		for _, wi := range ws {
			if hasElem(wi.TaskIDs, t) {
				return fmt.Errorf("Taskid %s has existed in workload %s", t, wi.ID)
			}
		}
	}
	return nil
}
