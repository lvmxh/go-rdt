package osgroup

import (
	"fmt"
	"strconv"
	"sync"

	"openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/resctrl"
	util "openstackcore-rdtagent/lib/util"
	. "openstackcore-rdtagent/util/rdtpool/base"
	. "openstackcore-rdtagent/util/rdtpool/osgroup/config"
)

var osGroupReserve = &Reserved{}
var once sync.Once

func GetOSGroupReserve() (Reserved, error) {
	var return_err error
	once.Do(func() {
		conf := NewConfig()
		osCPUbm, err := CpuBitmaps([]string{conf.CpuSet})
		if err != nil {
			return_err = err
			return
		}
		osGroupReserve.AllCPUs = osCPUbm

		level := syscache.GetLLC()
		syscaches, err := syscache.GetSysCaches(int(level))
		if err != nil {
			return_err = err
			return
		}

		// We though the ways number are same on all caches ID
		// FIXME if exception, fix it.
		ways, _ := strconv.Atoi(syscaches["0"].WaysOfAssociativity)
		if conf.CacheWays > uint(ways) {
			return_err = fmt.Errorf("The request OSGroup cache ways %d is larger than available %d.",
				conf.CacheWays, ways)
			return
		}

		schemata := map[string]*util.Bitmap{}
		osCPUs := map[string]*util.Bitmap{}

		for _, sc := range syscaches {
			bm, _ := CpuBitmaps([]string{sc.SharedCpuList})
			osCPUs[sc.Id] = osCPUbm.And(bm)
			if osCPUs[sc.Id].IsEmpty() {
				schemata[sc.Id], return_err = CacheBitmaps("0")
				if return_err != nil {
					return
				}
			} else {
				mask := strconv.FormatUint(1<<conf.CacheWays-1, 16)
				//FIXME (Shaohe) check RMD for the bootcheck.
				schemata[sc.Id], return_err = CacheBitmaps(mask)
				if return_err != nil {
					return
				}
			}
		}
		osGroupReserve.CPUsPerNode = osCPUs
		osGroupReserve.Schemata = schemata
	})

	return *osGroupReserve, return_err

}

func GetAvailableCaches(allres map[string]*resctrl.ResAssociation,
	reserve Reserved,
	cacheLevel string) map[string]*util.Bitmap {
	// FIXME (Shaohe) A central util to generate schemata Bitmap

	schemata := map[string]*util.Bitmap{}
	for k, _ := range reserve.Schemata {
		// "0-" + strconv.Itoa(int(ways))
		schemata[k], _ = CacheBitmaps(GetCosInfo().CbmMask)
	}
	for k, v := range allres {
		if k == "infra" || k == "." {
			continue
		}
		if sv, ok := v.Schemata[cacheLevel]; ok {
			for _, cv := range sv {
				k := strconv.Itoa(int(cv.Id))
				bm, _ := CacheBitmaps(cv.Mask)
				schemata[k] = schemata[k].Axor(bm)
			}
		}
	}
	return schemata
}
func SetOSGroup() error {
	reserve, err := GetOSGroupReserve()
	if err != nil {
		return err
	}

	allres := resctrl.GetResAssociation()
	osGroup := allres["."]
	org_bm, err := CpuBitmaps(osGroup.CPUs)
	if err != nil {
		return err
	}

	// NOTE (Shaohe), simpleness, brutal. Stolen CPUs from other groups.
	new_bm := org_bm.Or(reserve.AllCPUs)
	osGroup.CPUs = new_bm.ToString()

	level := syscache.GetLLC()
	target_lev := strconv.FormatUint(uint64(level), 10)
	cacheLevel := "L" + target_lev
	schemata := GetAvailableCaches(allres, reserve, cacheLevel)

	for i, v := range osGroup.Schemata[cacheLevel] {
		cacheId := strconv.Itoa(int(v.Id))
		if !reserve.CPUsPerNode[cacheId].IsEmpty() {
			// OSGroup is the first Group, use the edge cache ways.
			// FIXME (Shaohe), left or right cache ways, need to be check.
			conf := NewConfig()
			request, _ := CacheBitmaps(strconv.FormatUint(1<<conf.CacheWays-1, 16))
			// NOTE (Shaohe), simpleness, brutal. Reset Cache for OS Group,
			// even the cache is occupied by other group.
			available_ways := schemata[cacheId].Or(request)
			expect_ways := available_ways.ToBinStrings()[0]

			osGroup.Schemata[cacheLevel][i].Mask = strconv.FormatUint(1<<uint(len(expect_ways))-1, 16)
		} else {
			osGroup.Schemata[cacheLevel][i].Mask = GetCosInfo().CbmMask
		}
	}
	if err := osGroup.Commit("."); err != nil {
		return err
	}
	return nil
}
