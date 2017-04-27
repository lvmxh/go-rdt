package workload

// workload api objects to represent resources in RDTAgent

type RDTWorkLoad struct {
	// ID
	ID string
	// core ids, the work load run on top of cores/cpus
	CoreIDs []string `json:"core_ids"`
	// task ids, the work load's task ids
	TaskIDs []string `json:"task_ids"`
	// policy the workload want to apply
	Policys []string `json:"policys'`
	// algorithm  the workload want to apply
	Algorithms []string `json:"algorithms"`
	// Status
	Status string
}
