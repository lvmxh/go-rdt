package resctrl

import (
	"fmt"
	"github.com/denderello/task"
	"os"
	"path/filepath"
	"strings"
)

type ResctrlTask struct {
	TaskName string
	*ResAssociation
	RessSnapshot []ResAssociation
	Group        string
	Path         string
	Revert       bool // whether need to Revert after task faild
}

func (t ResctrlTask) Name() string {
	return t.TaskName
}

func (t ResctrlTask) Run() error {
	return nil
}

func (t ResctrlTask) Rollback() error {
	return nil
}

type ResctrlGroupTask struct {
	ResctrlTask
}

func (t ResctrlGroupTask) Run() error {
	return os.MkdirAll(t.Path, 0755)
}

func (t ResctrlGroupTask) Rollback() error {
	os.Remove(t.Path)
	return nil
}

type ResctrlCPUsTask struct {
	ResctrlTask
}

func (t ResctrlCPUsTask) Run() error {
	// Only write to cpus if admin specify cpu bit map
	// only commit a user deinfed cpus
	if t.CPUs != "" {
		return writeFile(t.Path, "cpus", t.CPUs)
	} else {
		return fmt.Errorf("Need to specify CPUs explicitly")
	}
}

func (t ResctrlCPUsTask) Rollback() error {
	if !t.Revert {
		return nil
	}
	// FIXME(Shaohe) need to revert the CPUs in all groups to the snapshort
	return nil
}

type ResctrlTasksTask struct {
	ResctrlTask
}

func (t ResctrlTasksTask) Run() error {
	// only commit a user deinfed group's task to sys fs
	if t.Group != "." && len(t.Tasks) > 0 {
		// write one task one time, or write will fail
		for _, v := range t.Tasks {
			err := writeFile(t.Path, "tasks", v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t ResctrlTasksTask) Rollback() error {
	if !t.Revert {
		return nil
	}
	// FIXME(Shaohe) need to revert the tasks in all groups to the snapshort
	return nil
}

type ResctrlSchemataTask struct {
	ResctrlTask
}

func (t ResctrlSchemataTask) Run() error {
	if len(t.Schemata) > 0 {
		schemata := make([]string, 0, 10)
		for k, v := range t.Schemata {
			str := make([]string, 0, 10)
			// resctrl require we have strict cache id order
			for cacheid := 0; cacheid < len(v); cacheid++ {
				for _, cos := range v {
					if uint8(cacheid) == cos.Id {
						str = append(str, fmt.Sprintf("%d=%s", cos.Id, cos.Mask))
						break
					}
				}
			}
			schemata = append(schemata, strings.Join([]string{k, strings.Join(str, ";")}, ":"))
		}
		data := strings.Join(schemata, "\n")
		err := writeFile(t.Path, "schemata", data)
		return err
	}
	return nil
}

func (t ResctrlSchemataTask) Rollback() error {
	// NOTE, do not need to revert the Schemata to the snapshort
	return nil
}

func TaskFlow(group string, r *ResAssociation, rs []ResAssociation) error {
	ts := []task.Task{}
	path := SysResctrl

	if strings.ToLower(group) != "default" && group != "." {
		path = filepath.Join(SysResctrl, group)
	}

	ct := ResctrlCPUsTask{ResctrlTask{"update-cpus", r, rs, group, path, true}}
	tt := ResctrlTasksTask{ResctrlTask{"update-tasks", r, rs, group, path, true}}
	st := ResctrlSchemataTask{ResctrlTask{"update-schemata", r, rs, group, path, true}}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		gt := ResctrlGroupTask{ResctrlTask{"creat-group", r, rs, group, path, true}}
		ct.Revert = false
		tt.Revert = false
		st.Revert = false
		ts = append(ts, gt)
	}

	ts = append(ts, []task.Task{ct, tt, st}...)
	q := task.NewQueue(ts)
	err := q.Start()
	if err != nil {
		// use log
		fmt.Println("Task queue failed: \n  ", err)
	}
	return err
}