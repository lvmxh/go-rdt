package rdtpool

import (
	"sync"

	"openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/cpu"
	util "openstackcore-rdtagent/lib/util"
	. "openstackcore-rdtagent/util/rdtpool/base"
	. "openstackcore-rdtagent/util/rdtpool/config"
)

// Workload can only get CPUs from this pool.
var cpuPoolPerCache = map[string]map[string]*util.Bitmap{
	"all":      map[string]*util.Bitmap{},
	"isolated": map[string]*util.Bitmap{}}
var cpuPoolOnce sync.Once

// helper function to get Reserved resource
func GetCPUPools() (map[string]map[string]*util.Bitmap, error) {
	var return_err error

	cpuPoolOnce.Do(func() {
		osconf := NewOSConfig()
		osCPUbm, err := CpuBitmaps([]string{osconf.CpuSet})
		if err != nil {
			return_err = err
			return
		}
		infraconf := NewInfraConfig()
		infraCPUbm, err := CpuBitmaps([]string{infraconf.CpuSet})
		if err != nil {
			return_err = err
			return
		}

		level := syscache.GetLLC()
		syscaches, err := syscache.GetSysCaches(int(level))
		if err != nil {
			return_err = err
			return
		}

		isocpu := cpu.IsolatedCPUs()
		var isolatedCPUbm *util.Bitmap
		if isocpu != "" {
			isolatedCPUbm, _ = CpuBitmaps([]string{cpu.IsolatedCPUs()})
		} else {
			isolatedCPUbm, _ = CpuBitmaps("Ox0")
		}

		for _, sc := range syscaches {
			bm, _ := CpuBitmaps([]string{sc.SharedCpuList})
			cpuPoolPerCache["all"][sc.Id] = bm.Axor(osCPUbm).Axor(infraCPUbm)
			cpuPoolPerCache["isolated"][sc.Id] = bm.Axor(osCPUbm).Axor(infraCPUbm).And(isolatedCPUbm)
		}
	})
	return cpuPoolPerCache, return_err
}
