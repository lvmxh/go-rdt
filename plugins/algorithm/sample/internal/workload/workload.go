package workload

// we can define RDTWorkLoad in model.
// Plugin can import this define.
type RDTWorkLoad struct {
	// ID
	ID string
	// core ids, the work load run on top of cores/cpus
	CoreIDs []string `json:"core_ids"`
	// task ids, the work load's task ids
	TaskIDs []string `json:"task_ids"`
	// policy the workload want to apply
	Policy string `json:"policy"`
	// Status
	Status string
	// Group
	Group []string `json:"group"`
	// CosName
	CosName string
}
