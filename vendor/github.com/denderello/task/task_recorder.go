package task

import "errors"

type TaskRecorder struct {
	TaskName           string
	RunSuccessful      bool
	RunExecuted        bool
	RollbackSuccessful bool
	RollbackExecuted   bool
}

func NewTaskRecorder(taskName string, runSuccessful, rollbackSuccessful bool) *TaskRecorder {
	return &TaskRecorder{
		TaskName:           taskName,
		RunSuccessful:      runSuccessful,
		RollbackSuccessful: rollbackSuccessful,
	}
}

func (t *TaskRecorder) Name() string {
	return t.TaskName
}

func (t *TaskRecorder) Run() error {
	t.RunExecuted = true

	if !t.RunSuccessful {
		return errors.New("error")
	}
	return nil
}

func (t *TaskRecorder) Rollback() error {
	t.RollbackExecuted = true

	if !t.RollbackSuccessful {
		return errors.New("error")
	}
	return nil
}
