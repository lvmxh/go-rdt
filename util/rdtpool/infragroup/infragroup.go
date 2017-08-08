package infragroup

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"

	"openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/proc"
	"openstackcore-rdtagent/lib/resctrl"
	util "openstackcore-rdtagent/lib/util"
	. "openstackcore-rdtagent/util/rdtpool/base"
	. "openstackcore-rdtagent/util/rdtpool/infragroup/config"
)

var groupName string = "infra"

var infraGroupReserve = &Reserved{}
var once sync.Once

func GetGlobTasks() []glob.Glob {
	conf := NewConfig()
	l := len(conf.Tasks)
	gs := make([]glob.Glob, l, l)
	for i, v := range conf.Tasks {
		g := glob.MustCompile(v)
		gs[i] = g
	}
	return gs
}

// NOTE (Shaohe) This group can be merged into GetOSGroupReserve
func GetInfraGroupReserve() (Reserved, error) {
	var return_err error
	once.Do(func() {
		conf := NewConfig()
		infraCPUbm, err := CpuBitmaps([]string{conf.CpuSet})
		if err != nil {
			return_err = err
			return
		}
		infraGroupReserve.AllCPUs = infraCPUbm

		level := syscache.GetLLC()
		syscaches, err := syscache.GetSysCaches(int(level))
		if err != nil {
			return_err = err
			return
		}

		// NOTE (Shaohe) here we do not guarantee OS and Infra Group will avoid overlap.
		// We can FIX it on bootcheek.
		// We though the ways number are same on all caches ID
		// FIXME if exception, fix it.
		ways, _ := strconv.Atoi(syscaches["0"].WaysOfAssociativity)
		if conf.CacheWays > uint(ways) {
			return_err = fmt.Errorf("The request InfraGroup cache ways %d is larger than available %d.",
				conf.CacheWays, ways)
			return
		}

		schemata := map[string]*util.Bitmap{}
		infraCPUs := map[string]*util.Bitmap{}

		for _, sc := range syscaches {
			bm, _ := CpuBitmaps([]string{sc.SharedCpuList})
			infraCPUs[sc.Id] = infraCPUbm.And(bm)
			if infraCPUs[sc.Id].IsEmpty() {
				schemata[sc.Id], return_err = CacheBitmaps("0")
				if return_err != nil {
					return
				}
			} else {
				// FIXME (Shaohe) We need to confirm the location of DDIO caches.
				// We Put on the left ways, opposite position of OS group cache ways.
				ways := uint(GetCosInfo().CbmMaskLen)
				mask := strconv.FormatUint((1<<conf.CacheWays-1)<<(ways-conf.CacheWays), 16)
				//FIXME (Shaohe) check RMD for the bootcheck.
				schemata[sc.Id], return_err = CacheBitmaps(mask)
				if return_err != nil {
					return
				}
			}
		}

		infraGroupReserve.CPUsPerNode = infraCPUs
		infraGroupReserve.Schemata = schemata
	})

	return *infraGroupReserve, return_err

}
func SetInfraGroup() error {
	conf := NewConfig()
	if conf == nil {
		return nil
	}

	reserve, err := GetInfraGroupReserve()
	if err != nil {
		return err
	}

	level := syscache.GetLLC()
	target_lev := strconv.FormatUint(uint64(level), 10)
	cacheLevel := "L" + target_lev
	ways := GetCosInfo().CbmMaskLen

	allres := resctrl.GetResAssociation()
	infraGroup, ok := allres[groupName]
	if !ok {
		infraGroup = resctrl.NewResAssociation()
		l := len(reserve.Schemata)
		infraGroup.Schemata[cacheLevel] = make([]resctrl.CacheCos, l, l)
	}
	infraGroup.CPUs = reserve.AllCPUs.ToString()

	for k, v := range reserve.Schemata {
		id, _ := strconv.Atoi(k)
		var mask string
		if !reserve.CPUsPerNode[k].IsEmpty() {
			mask = v.ToString()
		} else {
			mask = strconv.FormatUint(1<<uint(ways)-1, 16)
		}
		cc := resctrl.CacheCos{uint8(id), mask}
		infraGroup.Schemata[cacheLevel][id] = cc
	}

	gt := GetGlobTasks()
	tasks := []string{}
	ps := proc.ListProcesses()
	for k, v := range ps {
		for _, g := range gt {
			if g.Match(v.CmdLine) {
				tasks = append(tasks, k)
				log.Infof("Add task: %d to infra group. Command line: %s",
					v.Pid, v.CmdLine)
			}
		}
	}

	infraGroup.Tasks = append(infraGroup.Tasks, tasks...)

	if err := infraGroup.Commit(groupName); err != nil {
		return err
	}

	return nil
}
