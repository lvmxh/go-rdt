package task

import "container/list"

type Queue struct {
	taskList *list.List
}

func NewQueue(tasks []Task) *Queue {
	tl := list.New()
	for _, t := range tasks {
		tl.PushBack(t)
	}

	return &Queue{
		taskList: tl,
	}
}

func (q *Queue) Start() error {
	var runErr *runExecutionError
	var rollbackErr *rollbackExecutionError

	for te := q.taskList.Front(); te != nil; te = te.Next() {
		t := te.Value.(Task)
		err := t.Run()
		if err != nil {
			runErr = &runExecutionError{
				RootErr:       err,
				FailedElement: te,
			}
			break
		}
	}

	if runErr != nil {
		for te := runErr.FailedElement; te != nil; te = te.Prev() {
			t := te.Value.(Task)
			err := t.Rollback()
			if err != nil {
				rollbackErr = &rollbackExecutionError{
					RootErr:       err,
					FailedElement: te,
				}
				break
			}
		}

		return &QueueExecutionError{
			task:        runErr.FailedElement.Value.(Task),
			runErr:      runErr,
			rollbackErr: rollbackErr,
		}
	}

	return nil
}
