package rdtpool

import (
	"sync"

	"github.com/intel/rmd/lib/cache"
	"github.com/intel/rmd/lib/cpu"
	util "github.com/intel/rmd/lib/util"
	"github.com/intel/rmd/util/rdtpool/base"
	"github.com/intel/rmd/util/rdtpool/config"
)

// Workload can only get CPUs from this pool.
var cpuPoolPerCache = map[string]map[string]*util.Bitmap{
	"all":      map[string]*util.Bitmap{},
	"isolated": map[string]*util.Bitmap{}}
var cpuPoolOnce sync.Once

// GetCPUPools is helper function to get Reserved CPUs
func GetCPUPools() (map[string]map[string]*util.Bitmap, error) {
	var returnErr error

	cpuPoolOnce.Do(func() {
		osconf := config.NewOSConfig()
		osCPUbm, err := base.CPUBitmaps([]string{osconf.CPUSet})
		if err != nil {
			returnErr = err
			return
		}
		infraconf := config.NewInfraConfig()
		infraCPUbm, err := base.CPUBitmaps([]string{infraconf.CPUSet})
		if err != nil {
			returnErr = err
			return
		}

		level := syscache.GetLLC()
		syscaches, err := syscache.GetSysCaches(int(level))
		if err != nil {
			returnErr = err
			return
		}

		isocpu := cpu.IsolatedCPUs()
		var isolatedCPUbm *util.Bitmap
		if isocpu != "" {
			isolatedCPUbm, _ = base.CPUBitmaps([]string{cpu.IsolatedCPUs()})
		} else {
			isolatedCPUbm, _ = base.CPUBitmaps("Ox0")
		}

		for _, sc := range syscaches {
			bm, _ := base.CPUBitmaps([]string{sc.SharedCPUList})
			cpuPoolPerCache["all"][sc.ID] = bm.Axor(osCPUbm).Axor(infraCPUbm)
			cpuPoolPerCache["isolated"][sc.ID] = bm.Axor(osCPUbm).Axor(infraCPUbm).And(isolatedCPUbm)
		}
	})
	return cpuPoolPerCache, returnErr
}
