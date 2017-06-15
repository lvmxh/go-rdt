package db

import (
	"openstackcore-rdtagent/pkg/model/workload"
)

type DB interface {
	Initialize(dbname string)
	CreateWorkload(w *workload.RDTWorkLoad) error
	DeleteWorkload(w *workload.RDTWorkLoad) error
	GetAllWorkload() ([]workload.RDTWorkLoad, error)
	GetWorkloadById(id string) (workload.RDTWorkLoad, error)
}
