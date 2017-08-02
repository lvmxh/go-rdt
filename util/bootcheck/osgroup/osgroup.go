package osgroup

import (
	"fmt"
	"strconv"

	"openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/resctrl"
	util "openstackcore-rdtagent/lib/util"
	. "openstackcore-rdtagent/util/bootcheck/osgroup/config"
)

func SetOSGroup() error {
	conf := NewConfig()
	osCPUbm, err := CpuBitmaps([]string{conf.CpuSet})
	if err != nil {
		return err
	}
	if osCPUbm.IsEmpty() {
		return fmt.Errorf("must assign CPU set for OS group")
	}

	level := syscache.GetLLC()
	target_lev := strconv.FormatUint(uint64(level), 10)
	cacheLevel := "L" + target_lev
	syscaches, err := syscache.GetSysCaches(int(level))
	if err != nil {
		return err
	}

	// We though the ways number are same on all caches ID
	// FIXME if exception, fix it.
	ways, _ := strconv.Atoi(syscaches["0"].WaysOfAssociativity)
	if conf.CacheWays > uint(ways) {
		return fmt.Errorf("The request OSGroup cache ways %d is larger than available %d.",
			conf.CacheWays, ways)
	}

	osCPUs := map[string]*util.Bitmap{}
	resaall := resctrl.GetResAssociation()

	osGroup := resaall["."]
	org_bm, err := CpuBitmaps(osGroup.CPUs)
	if err != nil {
		return err
	}

	// NOTE (Shaohe), simpleness, brutal. Stolen CPUs from other groups.
	new_bm := org_bm.Or(osCPUbm)
	osGroup.CPUs = new_bm.ToString()

	for _, sc := range syscaches {
		bm, _ := CpuBitmaps([]string{sc.SharedCpuList})
		osCPUs[sc.Id] = osCPUbm.And(bm)
	}

	for i, v := range osGroup.Schemata[cacheLevel] {
		if !osCPUs[strconv.Itoa(int(v.Id))].IsEmpty() {
			// OSGroup is the first Group, use the edge cache ways.
			// FIXME (Shaohe), left or right cache ways, need to be check.
			osGroup.Schemata[cacheLevel][i].Mask = strconv.FormatUint(1<<conf.CacheWays-1, 16)
		} else {
			// assume the caches ways less than 32.
			osGroup.Schemata[cacheLevel][i].Mask = strconv.FormatUint(1<<uint(ways)-1, 16)
		}
	}

	osGroup.Commit(".")
	return nil
}
