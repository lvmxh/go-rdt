package workload

// workload api objects to represent resources in RDTAgent

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"openstackcore-rdtagent/lib/cpu"
	"openstackcore-rdtagent/lib/proc"
	"openstackcore-rdtagent/lib/resctrl"
	libutil "openstackcore-rdtagent/lib/util"
	"openstackcore-rdtagent/model/cache"
	"openstackcore-rdtagent/model/policy"
	modelutil "openstackcore-rdtagent/model/util"
)

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

func (w *RDTWorkLoad) Enforce() error {
	// FIXME First check CAT is enabled

	if len(w.TaskIDs) <= 0 && len(w.CoreIDs) <= 0 {
		return fmt.Errorf("No task or core id specified")
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

	cpubitmap, err := libutil.GenCpuResString(w.CoreIDs, cpunum)
	if err != nil {
		return err
	}

	// TODO(eliqiao): if no group sepcify, isolated = true
	// please refine this in later version
	p, err := policy.GetPolicy("broadwell", w.Policy)
	if err != nil {
		return err
	}

	resaall := resctrl.GetResAssociation()

	if len(w.Group) > 2 {
		// FIXME what if more than 3 groups
		return fmt.Errorf("Can not specified more then 2 group list name")
	}

	base_grp, new_grp, sub_grp := getGroupNames(w, resaall)

	if base_grp == "" {
		// log group information
		return fmt.Errorf("Faild to find a suitable group")
	}

	peakusage, err := strconv.Atoi(p["peakusage"])

	if err != nil {
		return err
	}

	targetResAss, err := createOrGetResAss(resaall, base_grp, new_grp, sub_grp, uint32(peakusage*1024))
	if err != nil {
		// log
		return err
	}

	targetResAss.Tasks = append(targetResAss.Tasks, w.TaskIDs...)
	// FIXME need to check if we need to change cpubitmap
	targetResAss.Cpus = cpubitmap

	if base_grp != new_grp && base_grp != "." {
		new_grp = base_grp + "-" + new_grp
	}

	err = targetResAss.Commit(new_grp)
	if err != nil {
		// log error
		return err
	}

	if base_grp == "." {
		if err = resaall["."].Commit("."); err != nil {
			return err
		}
	}

	//foo(resaall["."], 1, true)
	if len(w.Group) == 0 {
		w.Group = append(w.Group, new_grp)
	}

	w.CosName = new_grp

	return nil
}

// Release Cos
func (w *RDTWorkLoad) Release() error {
	resaall := resctrl.GetResAssociation()

	r, ok := resaall[w.CosName]

	if !ok {
		return nil
	}

	r.Tasks = modelutil.SubtractStringSlice(r.Tasks, w.TaskIDs)

	// safely remove resource group if no tasks and cpu bit map is empty
	if len(r.Tasks) < 1 {
		log.Printf("Remove resource group: %s", w.CosName)
		if err := r.Remove(w.CosName); err != nil {
			return err
		}
		// TODO (eliqiao): try compensate cbm to default group
	}

	// remove workload task ids from resource group
	if len(w.TaskIDs) > 0 {
		if err := resctrl.RemoveTasks(w.TaskIDs); err != nil {
			log.Printf("Ignore Error while remove tasks %s", err)
			return nil
		}
	}

	return nil
}

// return base group name, new group name, sub group name list.
// e.g.
// CG1 L3:0=ffff;1=ffff
// CG1-SUB1 L3:0=f;1=f
// CG2 L3:0=f0000;1=f0000
//
// if w.Group is ["CG1", "SUB2"]
// getGroupNames will return CG1, SUB2, [CG1-SUB1]
func getGroupNames(w *RDTWorkLoad, m map[string]*resctrl.ResAssociation) (b, n string, s []string) {
	var new_grp string
	var base_grp string
	sub_grp := []string{}
	// no group specify
	if len(w.Group) == 0 {
		if len(w.TaskIDs) > 0 {
			// use the first task id as gname
			new_grp = w.TaskIDs[0]
		} else {
			// FIXME generate a better group name
			new_grp = w.ID
		}
		return ".", new_grp, []string{}
	}

	if len(w.Group) == 1 {
		_, ok := m[w.Group[0]]
		if ok {
			// find existed group
			// new group and base group are same
			return w.Group[0], w.Group[0], []string{}
		} else {
			// doesn't find one, create a new one
			return ".", w.Group[0], []string{}
		}
	}

	if len(w.Group) == 2 {
		_, ok1 := m[w.Group[0]]
		_, ok2 := m[w.Group[1]]
		if ok1 && ok2 {
			// FIXME error
			return "", "", []string{}
		}
		if !ok1 && !ok2 {
			// FIXME error
			return "", "", []string{}
		}

		if ok1 {
			base_grp = w.Group[0]
			new_grp = w.Group[1]
		} else {
			base_grp = w.Group[1]
			new_grp = w.Group[2]
		}
		for g, _ := range m {
			// sub group names like base-sub
			if strings.HasPrefix(g, base_grp+"-") {
				sub_grp = append(sub_grp, g)
			}
		}
		return base_grp, new_grp, sub_grp
	}
	// error
	return "", "", []string{}
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
func updateMask(basemask, size, unit, offset uint32, consume bool) (newbasemask, newmask uint32) {

	newmask = getMaskBySize(size, unit)
	if newmask >= basemask {
		// todo log
		return basemask, 0
	}

	newmask = newmask << offset

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
// return a Resassociation with proper mask set
func createOrGetResAss(r map[string]*resctrl.ResAssociation, base_grp, new_grp string, sub_grp []string, size uint32) (t resctrl.ResAssociation, err error) {
	if base_grp == new_grp {
		return *r[base_grp], nil
	}
	for _, sg := range sub_grp {
		if base_grp+"-"+new_grp == sg {
			// new_grp has existed
			return *r[sg], nil
		}
	}
	// consider move consume checking to createNewResassociation
	if base_grp == "." {
		// sub_grp should be empty if the base group is "."
		// or that should be an internal error.
		return createNewResassociation(r, ".", size, true, []string{})
	}
	return createNewResassociation(r, base_grp, size, false, sub_grp)
}

// size is in B
// return a new Resassociation based on the given resctrl.ResAssociation
func createNewResassociation(r map[string]*resctrl.ResAssociation, base string, size uint32, consume bool, sub_grp []string) (t resctrl.ResAssociation, err error) {
	rdtinfo := resctrl.GetRdtCosInfo()
	cacheinfo := &cache.CacheInfos{}
	// Fixme Upper layer should pass a cache level parameter
	cacheinfo.GetByLevel(3)

	// loop for each level 3 cache to construct new resassociation
	newResAss := resctrl.ResAssociation{}
	newResAss.Schemata = make(map[string][]resctrl.CacheCos)

	for cattype, res := range r[base].Schemata {
		catinfo := rdtinfo[strings.ToLower(cattype)]
		// CbmMask is in hex
		cbmlen := modelutil.CbmLen(catinfo.CbmMask)
		// construct ResAssociation for each cache id
		for i, c := range cacheinfo.Caches {
			// compute sub_grp's offset for the i(th) 'cattype'
			offset := calculateOffset(r, sub_grp, cattype, i)
			unit := c.TotalSize / uint32(cbmlen)
			newbasemask, newmask := updateMask(res[i].Mask, size, unit, offset, consume)
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

// Calculate offset for the pos'th cache of cattype based on sub_grp
// e.g.
// sub_grp = [base-sub1]
// base-sub1: L3:0=f;1=1
// calculateOffset(r, sub_grp, L3, 0) = 4
// calculateOffset(r, sub_grp, L3, 1) = 1
func calculateOffset(r map[string]*resctrl.ResAssociation, sub_grp []string, cattype string, pos uint32) uint32 {
	var biggestMask, offset uint32
	biggestMask = 0

	for _, g := range sub_grp {
		if biggestMask < r[g].Schemata[cattype][pos].Mask {
			biggestMask = r[g].Schemata[cattype][pos].Mask
		}
	}

	for offset = 0; biggestMask > 0; biggestMask = biggestMask >> 1 {
		offset += 1
	}
	return offset
}
