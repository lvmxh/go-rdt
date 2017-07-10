package workload

// workload api objects to represent resources in RDTAgent

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"sync"

	libcache "openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/cpu"
	"openstackcore-rdtagent/lib/proc"
	"openstackcore-rdtagent/lib/resctrl"
	libutil "openstackcore-rdtagent/lib/util"

	"openstackcore-rdtagent/model/cache"
	"openstackcore-rdtagent/model/policy"
	modelutil "openstackcore-rdtagent/model/util"
)

// global lock for when doing enforce/update/release for a workload.
// This is a simple way to control RDAgent to access resctrl one
// goroutine one time
var l sync.Mutex

const (
	Successful = "Successful"
	Failed     = "Failed"
	Invalid    = "Invalid"
	None       = "None"
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

func (w *RDTWorkLoad) Validate() error {
	if len(w.TaskIDs) <= 0 && len(w.CoreIDs) <= 0 {
		return fmt.Errorf("No task or core id specified")
	}

	// Firstly verify the task.
	ps := proc.ListProcesses()
	for _, task := range w.TaskIDs {
		if _, ok := ps[task]; !ok {
			return fmt.Errorf("The task: %s does not exist", task)
		}
	}
	// user don't need to provide group name anymore, if we configured
	// infra_group, then let RDAgent append the group name
	// e.g. w.Group = append(w.Group, "infra")
	// e.g. w.Group = append(w.Group, "w.name")
	if len(w.Group) > 2 {
		return fmt.Errorf("Can not specified more then 2 group list name")
	}

	return nil
}

func (w *RDTWorkLoad) Enforce() error {
	if err := w.Validate(); err != nil {
		log.Errorf("Failed to validate workload %s, error: %s", w.ID, err)
		w.Status = Invalid
		return err
	}

	w.Status = None
	// FIXME cpunum can be global
	cpunum, err := cpu.HostCpuNum()
	if err != nil {
		return err
	}

	cpubitmap, err := libutil.GenCpuResString(w.CoreIDs, cpunum)
	if err != nil {
		return err
	}

	// status will be updated to successful if no errors
	w.Status = Failed

	l.Lock()
	defer l.Unlock()
	resaall := resctrl.GetResAssociation()

	base_grp, new_grp, sub_grp := getGroupNames(w, resaall)

	if base_grp == "" {
		// log group information
		return fmt.Errorf("Faild to find a suitable group")
	}

	log.Debugf("base group %s, new group %s, sub group %v", base_grp, new_grp, sub_grp)

	p, err := policy.GetPolicy("broadwell", w.Policy)
	if err != nil {
		return err
	}

	ways, err := strconv.Atoi(p["MaxCache"])
	if err != nil {
		return err
	}

	targetResAss, err := createOrGetResAss(resaall, base_grp, new_grp, sub_grp, uint32(ways))
	if err != nil {
		log.Errorf("Error while try to create resource group for workload %s", w.ID)
		return err
	}

	targetResAss.Tasks = append(targetResAss.Tasks, w.TaskIDs...)
	targetResAss.CPUs = cpubitmap

	if base_grp != new_grp && base_grp != "." {
		new_grp = base_grp + "-" + new_grp
	}

	if err = targetResAss.Commit(new_grp); err != nil {
		log.Errorf("Error while try to commit resource group for workload %s, group name %s", w.ID, new_grp)
		return err
	}

	if base_grp == "." {
		if err = resaall["."].Commit("."); err != nil {
			log.Errorf("Error while try to commit resource group for default group")
			resctrl.DestroyResAssociation(new_grp)
			return err
		}
	}

	if len(w.Group) == 0 {
		w.Group = append(w.Group, new_grp)
	}

	w.CosName = new_grp
	w.Status = Successful
	return nil
}

// Release Cos
func (w *RDTWorkLoad) Release() error {
	l.Lock()
	defer l.Unlock()

	resaall := resctrl.GetResAssociation()

	r, ok := resaall[w.CosName]

	if !ok {
		return nil
	}

	r.Tasks = modelutil.SubtractStringSlice(r.Tasks, w.TaskIDs)

	// safely remove resource group if no tasks and cpu bit map is empty
	if len(r.Tasks) < 1 {
		log.Printf("Remove resource group: %s", w.CosName)
		if err := resctrl.DestroyResAssociation(w.CosName); err != nil {
			return err
		}
		delete(resaall, w.CosName)
		newDefaultGrp := calculateDefaultGroup(resaall, []string{"."}, true)
		newDefaultGrp.Commit(".")
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

// return a Resassociation with proper mask set
func createOrGetResAss(r map[string]*resctrl.ResAssociation, base_grp, new_grp string, sub_grp []string, ways uint32) (t resctrl.ResAssociation, err error) {
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
		return createNewResassociation(r, ".", ways, true, []string{})
	}
	return createNewResassociation(r, base_grp, ways, false, sub_grp)
}

// return a new Resassociation based on the given resctrl.ResAssociation
func createNewResassociation(r map[string]*resctrl.ResAssociation, base string, ways uint32, consume bool, sub_grp []string) (t resctrl.ResAssociation, err error) {
	cacheinfo := &cache.CacheInfos{}
	cacheinfo.GetByLevel(libcache.GetLLC())

	baseRes := r[base]
	if base == "." {
		// if infra group are created, should be added it to ignore group.
		baseRes = calculateDefaultGroup(r, []string{"."}, false)
		r["."] = baseRes
	}

	// loop for each level 3 cache to construct new resassociation
	newResAss := resctrl.ResAssociation{}
	newResAss.Schemata = make(map[string][]resctrl.CacheCos)

	for cattype, res := range baseRes.Schemata {
		// construct ResAssociation for each cache id
		for i, _ := range cacheinfo.Caches {
			// compute sub_grp's offset for the i(th) 'cattype'
			offset := calculateOffset(r, sub_grp, cattype, i)

			// len is not so important, we don't want to query cbm_mask every time
			// we new a bitmap, this is too much time costing, later we need to load
			// Len(cbm_mask) as a global variable
			bmbase, _ := libutil.NewBitmap(20, res[i].Mask)
			newbm := bmbase.GetConnectiveBits(ways, offset, true)

			if newbm.IsEmpty() {
				return newResAss, fmt.Errorf("Not enough cache can be allocated")
			}

			if consume {
				bmbase = bmbase.Xor(newbm)
			}

			tmpbm := bmbase.MaxConnectiveBits()
			res[i].Mask = tmpbm.ToString()
			newcos := resctrl.CacheCos{Id: uint8(i), Mask: newbm.ToString()}

			newResAss.Schemata[cattype] = append(newResAss.Schemata[cattype], newcos)
			log.Debugf("Newly created Mask for Cache %d is %s", i, newcos.Mask)
			log.Debugf("Default Mask for Cache %d is %s", i, res[i].Mask)
		}
	}
	return newResAss, nil
}

// calculate a new default group which can be consumed by
// subtract bit mask in existed resource group
// e.g.
// r = ['.': [res1], 'group1': 'res2', 'infra': 'bla']
// ignore_grp = ['.', 'infra']
// calculateDefaultGroup will create a new resctrl.ResAssociation, the mask
// value is calculated by default mask subtract group1's schemata.
func calculateDefaultGroup(r map[string]*resctrl.ResAssociation, ignore_grp []string, consecutive bool) *resctrl.ResAssociation {

	defaultGrp := r["."]

	for _, v := range ignore_grp {
		delete(r, v)
	}

	// FIXME rdtinfo could be a global variable
	rdtinfo := resctrl.GetRdtCosInfo()
	newRes := new(resctrl.ResAssociation)
	newRes.Schemata = make(map[string][]resctrl.CacheCos)

	for t, schemata := range defaultGrp.Schemata {
		catinfo := rdtinfo[strings.ToLower(t)]
		newRes.Schemata[t] = make([]resctrl.CacheCos, 0, 10)
		for id, v := range schemata {
			// len is not so important, we don't want to query cbm_mask every time
			// we new a bitmap, this is too much time costing, later we need to load
			// Len(cbm_mask) as a global variable
			bm, _ := libutil.NewBitmap(20, catinfo.CbmMask)
			// loop for all groups
			for _, g := range r {
				// len is not so important, we don't want to query cbm_mask every time
				// we new a bitmap, this is too much time costing, later we need to load
				// Len(cbm_mask) as a global variable
				gbm, _ := libutil.NewBitmap(20, g.Schemata[t][v.Id].Mask)
				bm = bm.Xor(gbm)
			}

			var newcbm string

			if consecutive {
				tmpbm := bm.MaxConnectiveBits()
				newcbm = tmpbm.ToString()
			} else {
				newcbm = bm.ToString()
			}
			cacheCos := &resctrl.CacheCos{uint8(id), newcbm}
			newRes.Schemata[t] = append(newRes.Schemata[t], *cacheCos)

			log.Debugf("New default Mask for Cache %d is %s", cacheCos.Id, newcbm)
		}
	}

	return newRes
}

// Calculate offset for the pos'th cache of cattype based on sub_grp
// e.g.
// sub_grp = [base-sub1]
// base-sub1: L3:0=f;1=1
// calculateOffset(r, sub_grp, L3, 0) = 4
// calculateOffset(r, sub_grp, L3, 1) = 1
func calculateOffset(r map[string]*resctrl.ResAssociation, sub_grp []string, cattype string, pos uint32) uint32 {
	// len is not so important, we don't want to query cbm_mask every time
	// we new a bitmap, this is too much time costing, later we need to load
	// Len(cbm_mask) as a global variable
	bm0, _ := libutil.NewBitmap(20, "")
	for _, g := range sub_grp {
		b, _ := libutil.NewBitmap(20, r[g].Schemata[cattype][pos].Mask)
		bm0 = bm0.Or(b)
	}
	if bm0.IsEmpty() {
		return 0
	} else {
		//return bm0.Maximum()
		// TODO(shaohe): need to implement bm0.Maximum()
		return 0
	}
}
