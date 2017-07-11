package task

import (
	"container/list"
	"fmt"
)

type runExecutionError struct {
	RootErr       error
	FailedElement *list.Element
}

func (e *runExecutionError) Error() string {
	return fmt.Sprintf("Run failed: %s", e.RootErr)
}

type rollbackExecutionError struct {
	RootErr       error
	FailedElement *list.Element
}

func (e *rollbackExecutionError) Error() string {
	return fmt.Sprintf("Rollback failed: %s", e.RootErr)
}

type QueueExecutionError struct {
	task        Task
	runErr      *runExecutionError
	rollbackErr *rollbackExecutionError
}

func (e *QueueExecutionError) Error() string {
	msg := fmt.Sprintf("Task %s failed.\n\t%s", e.task.Name(), e.runErr.Error())
	if e.rollbackErr != nil {
		msg = fmt.Sprintf("%s\n\t%s", msg, e.rollbackErr.Error())
	}
	return msg
}
