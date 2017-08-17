package rdtpool

import (
	"fmt"
	"strconv"
	"sync"

	"openstackcore-rdtagent/lib/cache"
	util "openstackcore-rdtagent/lib/util"
	. "openstackcore-rdtagent/util/rdtpool/base"
	. "openstackcore-rdtagent/util/rdtpool/config"
)

var cachePoolReserved = make(map[string]*Reserved, 0)
var cachePoolOnce sync.Once

func GetCachePoolLayout() (map[string]*Reserved, error) {
	var return_err error
	cachePoolOnce.Do(func() {
		poolConf := NewCachePoolConfig()
		osConf := NewOSConfig()
		ways := GetCosInfo().CbmMaskLen

		if osConf.CacheWays+poolConf.Guarantee+poolConf.Besteffort > uint(ways) {
			return_err = fmt.Errorf(
				"Error config: Guarantee + Besteffort + OS reserved ways should be less or equal to %d.", ways)
			return
		}

		if poolConf.Besteffort < poolConf.Shared {
			return_err = fmt.Errorf(
				"Error config: Shared ways %d should be less or equal to Besteffort ways %d.",
				poolConf.Shared, poolConf.Besteffort)
			return
		}

		// set layout for cache pool
		level := syscache.GetLLC()
		syscaches, err := syscache.GetSysCaches(int(level))
		osCPUbm, err := CpuBitmaps([]string{osConf.CpuSet})

		if err != nil {
			return_err = err
			return
		}

		// FIXME: improve this logic
		if poolConf.Guarantee > 0 {
			schemata := map[string]*util.Bitmap{}
			osCPUs := map[string]*util.Bitmap{}

			for _, sc := range syscaches {
				bm, _ := CpuBitmaps([]string{sc.SharedCpuList})
				osCPUs[sc.Id] = osCPUbm.And(bm)

				wayCandidate := 1<<poolConf.Guarantee - 1
				// no os group on this cache id
				if !osCPUs[sc.Id].IsEmpty() {
					wayCandidate = wayCandidate << osConf.CacheWays
				}
				mask := strconv.FormatUint(uint64(wayCandidate), 16)
				schemata[sc.Id], return_err = CacheBitmaps(mask)
				if return_err != nil {
					return
				}
			}
			guaranteePoolReserved := &Reserved{}
			guaranteePoolReserved.Schemata = schemata

			cachePoolReserved[Guarantee] = guaranteePoolReserved
		}

		// FIXME: improve this logic
		if poolConf.Besteffort > 0 {
			schemata := map[string]*util.Bitmap{}
			osCPUs := map[string]*util.Bitmap{}

			for _, sc := range syscaches {
				bm, _ := CpuBitmaps([]string{sc.SharedCpuList})
				osCPUs[sc.Id] = osCPUbm.And(bm)

				wayCandidate := 1<<poolConf.Besteffort - 1
				wayCandidate = wayCandidate << poolConf.Guarantee
				// no os group on this cache id
				if !osCPUs[sc.Id].IsEmpty() {
					wayCandidate = wayCandidate << osConf.CacheWays
				}
				mask := strconv.FormatUint(uint64(wayCandidate), 16)
				schemata[sc.Id], return_err = CacheBitmaps(mask)
				if return_err != nil {
					return
				}
			}
			besteffortPoolReserved := &Reserved{}
			besteffortPoolReserved.Schemata = schemata

			cachePoolReserved[Besteffort] = besteffortPoolReserved
		}

		// FIXME: improve this logic
		if poolConf.Shared > 0 {
			schemata := map[string]*util.Bitmap{}
			osCPUs := map[string]*util.Bitmap{}

			for _, sc := range syscaches {
				bm, _ := CpuBitmaps([]string{sc.SharedCpuList})
				osCPUs[sc.Id] = osCPUbm.And(bm)
				wayCandidate := 1<<poolConf.Shared - 1
				offset := (poolConf.Besteffort - poolConf.Shared) / 2
				wayToMove := poolConf.Guarantee + poolConf.Besteffort - poolConf.Shared - offset
				wayCandidate = wayCandidate << wayToMove
				if !osCPUs[sc.Id].IsEmpty() {
					wayCandidate = wayCandidate << osConf.CacheWays
				}

				mask := strconv.FormatUint(uint64(wayCandidate), 16)
				schemata[sc.Id], return_err = CacheBitmaps(mask)
				if return_err != nil {
					return
				}
			}

			sharedPoolReserved := &Reserved{Quota: poolConf.MaxAllowedShared}
			sharedPoolReserved.Schemata = schemata

			cachePoolReserved[Shared] = sharedPoolReserved
		}
	})

	return cachePoolReserved, return_err

}
