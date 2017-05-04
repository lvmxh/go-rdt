package workload

// workload api objects to represent resources in RDTAgent

import (
	"fmt"
	"strings"

	"openstackcore-rdtagent/lib/cpu"
	"openstackcore-rdtagent/lib/proc"
	"openstackcore-rdtagent/lib/resctrl"
	libutil "openstackcore-rdtagent/lib/util"
	"openstackcore-rdtagent/pkg/model/cache"
	"openstackcore-rdtagent/pkg/model/policy"
)

type RDTWorkLoad struct {
	// ID
	ID string
	// core ids, the work load run on top of cores/cpus
	CoreIDs []string `json:"core_ids"`
	// task ids, the work load's task ids
	TaskIDs []string `json:"task_ids"`
	// policy the workload want to apply
	Policy string `json:"policy'`
	// Status
	Status string
	// Group
	Group []string `json:"group"`
}

func (w *RDTWorkLoad) Enforce() error {

	if len(w.TaskIDs) < 0 {
		return fmt.Errorf("No task ids specified")
	}

	// Firstly verify the task.
	ps := proc.ListProcesses()
	for _, task := range w.TaskIDs {
		if _, ok := ps[task]; !ok {
			return fmt.Errorf("The workload: %s does not exit", task)
		}
	}

	// FIXME cpunum can be global
	cpunum, err := cpu.HostCpuNum()
	if err != nil {
		return err
	}

	cpubitmap, err := libutil.GenerateBitMap(w.CoreIDs, cpunum)
	if err != nil {
		return err
	}

	// TODO(eliqiao): if no group sepcify, isolated = true
	// please refine this in later version
	isolated := true
	p, err := policy.GetPolicy("haswell")
	if err != nil {
		return err
	}

	var pt policy.PolicyType
	var gname string

	switch w.Policy {
	case "gold":
		pt = p.Gold
	case "silver":
		pt = p.Silver
	case "copper":
		pt = p.Copper
	}

	fmt.Println(pt.Size)

	resaall := resctrl.GetResAssociation()

	if len(w.Group) > 0 {
		// TODO (eliqiao):
		// in this branch we may need to create overlap COS
		// or add tasks to an existed COS
		for _, g := range w.Group {
			fmt.Println(g)
			for osg, _ := range resaall {
				fmt.Println(osg)
				if g == osg {
					break
				}
			}
		}
	} else {
		// create a new group with the first task id
		// TODO(eliqiao): verify task id is valid
		gname = w.TaskIDs[0]
	}

	if gname != "" {
		// in this branch we convert the size to a cos base on default
		// schemata then commit it to sysfs
		newResAss, err := createNewResAss(resaall["."], pt.Size*1024, isolated)
		if err != nil {
			// log
			return err
		}
		newResAss.Tasks = w.TaskIDs
		newResAss.Cpus = cpubitmap
		err = newResAss.Commit(gname)
		if err != nil {
			// log error
			return err
		}
		if isolated == true {
			if err = resaall["."].Commit("."); err != nil {
				// log error
				return err
			}
		}
		w.Group = append(w.Group, gname)
	}

	return nil
}

// calculate the mask based on the size
func getMaskBySize(size, unit uint32) (ret uint32) {

	if size <= 0 {
		return 0
	}

	bitcounts := size / unit

	if size%unit > 0 || bitcounts == 0 {
		bitcounts += 1
	}

	return (1 << (bitcounts)) - 1
}

// update the mask based on base mask, and consume it
// return net base mask and new mask, if error the new mask is 0
func updateMask(basemask, size, unit uint32, consume bool) (newbasemask, newmask uint32) {

	newmask = getMaskBySize(size, unit)
	if newmask >= basemask {
		// todo log
		return basemask, 0
	}

	// start from the most right place
	for newmask < basemask {
		if newmask&basemask == newmask {
			break
		} else {
			newmask = newmask << 1
		}
	}

	if consume == true {
		newbasemask = basemask - newmask
	}
	return newbasemask, newmask
}

// size is in B
// return a new Resassociation based on the given resctrl.ResAssociation
func createNewResAss(r *resctrl.ResAssociation, size uint32, consume bool) (t resctrl.ResAssociation, err error) {
	rdtinfo := resctrl.GetRdtCosInfo()
	cacheinfo := &cache.CacheInfos{}
	// Fixme Upper layer should pass a cache level parameter
	cacheinfo.GetByLevel(3)

	// loop for each level 3 cache to construct new resassociation
	newResAss := resctrl.ResAssociation{}
	newResAss.Schemata = make(map[string][]resctrl.CacheCos)

	for cattype, res := range r.Schemata {
		catinfo := rdtinfo[strings.ToLower(cattype)]
		// CbmMask is in hex
		cbmlen := len(catinfo.CbmMask) * 4
		// construct ResAssociation for each cache id
		for i, c := range cacheinfo.Caches {
			unit := c.TotalSize / uint32(cbmlen)
			newbasemask, newmask := updateMask(res[i].Mask, size, unit, consume)
			res[i].Mask = newbasemask
			if newmask == 0 {
				return newResAss, fmt.Errorf("Not enough cache can be allocated")
			}
			newcos := resctrl.CacheCos{Id: uint8(i), Mask: newmask}
			newResAss.Schemata[cattype] = append(newResAss.Schemata[cattype], newcos)
		}
	}
	return newResAss, nil
}
