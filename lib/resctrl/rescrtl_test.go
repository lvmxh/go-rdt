// +build linux

package resctrl

import (
	"errors"
	"fmt"
	"github.com/denderello/task"
	"testing"
)

func TestGetResAssociation(t *testing.T) {
	ress := GetResAssociation()
	for name, res := range ress {
		if name == "CG1" {
			fmt.Println(name)
			fmt.Println(res)
			fmt.Println(res.Schemata["L3CODE"])
		}
	}
}

func TestGetRdtCosInfo(t *testing.T) {

	infos := GetRdtCosInfo()
	for name, info := range infos {
		fmt.Println(name)
		fmt.Println(info)
	}
}

type ResctrlTasksTask struct {
	TaskName string
	Tasks    []string
}

func (t ResctrlTasksTask) Name() string {
	return t.TaskName
}

func (t ResctrlTasksTask) Run() error {
	fmt.Printf("Running task %s\n", t.Name())
	fmt.Println("Write tasks", t.Tasks, "to file")
	return nil
}

// This also can be an empty function.
func (t ResctrlTasksTask) Rollback() error {
	fmt.Printf("Ignore task %s Rolling back!\n", t.Name())
	return nil
}

type ResctrlCpuTask struct {
	TaskName string
	CPU      string
}

func (t ResctrlCpuTask) Name() string {
	return t.TaskName
}

func (t ResctrlCpuTask) Run() error {
	fmt.Printf("Running task %s\n", t.Name())
	fmt.Println("Write CPUs ", t.CPU, "to file")
	return nil
}

func (t ResctrlCpuTask) Rollback() error {
	fmt.Printf("Rolling back task %s\n", t.Name())
	return nil
}

type ResctrlSchemataTask struct {
	TaskName string
	Schemata string
	Err      bool
}

func (t ResctrlSchemataTask) Name() string {
	return t.TaskName
}

func (t ResctrlSchemataTask) Run() error {

	if t.Err {
		fmt.Printf("write schemata %s failed.\n", t.Schemata)
		return errors.New("error")
	} else {
		fmt.Printf("write schemata %s successfully.\n", t.Schemata)
	}
	return nil
}

func (t ResctrlSchemataTask) Rollback() error {
	fmt.Printf("Rolling back task %s\n", t.Name())
	return nil
}

func TestUpdateNewGroupFailed(t *testing.T) {
	ts := []task.Task{
		ResctrlCpuTask{"task-CPU", "F0,FFFF0000"},
		ResctrlTasksTask{"task-Tasks", []string{"1234", "5678", "4859"}},
		ResctrlSchemataTask{"task-schemata", "L3:0=7ff;1=7ff", true},
		// we can add other task into this slice.
	}

	q := task.NewQueue(ts)
	err := q.Start()
	if err != nil {
		fmt.Println("Task queue failed: \n  ", err)
	}
}

func TestUpdateNewGroupSuccessfully(t *testing.T) {
	ts := []task.Task{
		ResctrlCpuTask{"task-CPU", "F0,FFFF0000"},
		ResctrlTasksTask{"task-Tasks", []string{"1234", "5678", "4859"}},
		ResctrlSchemataTask{"task-schemata", "L3:0=7ff;1=7ff", false},
		// we can add other task into this slice.
	}

	q := task.NewQueue(ts)
	err := q.Start()
	if err != nil {
		fmt.Println("Task queue failed: \n  ", err)
	}
}
