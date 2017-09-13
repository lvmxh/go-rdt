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

// helper function to get Reserved resource
func getReservedCache(
	wayCandidate int,
	wayOffset, osCacheWays uint,
	osCPUbm *util.Bitmap,
	sysc map[string]syscache.SysCache) (*Reserved, error) {

	r := &Reserved{}

	schemata := map[string]*util.Bitmap{}
	osCPUs := map[string]*util.Bitmap{}
	var err error

	for _, sc := range sysc {
		wc := wayCandidate
		bm, _ := CpuBitmaps([]string{sc.SharedCpuList})
		osCPUs[sc.Id] = osCPUbm.And(bm)
		// no os group on this cache id
		if !osCPUs[sc.Id].IsEmpty() {
			wc = wc << osCacheWays
		}
		wc = wc << wayOffset
		mask := strconv.FormatUint(uint64(wc), 16)
		schemata[sc.Id], err = CacheBitmaps(mask)
		if err != nil {
			return r, err
		}
	}

	r.Schemata = schemata
	return r, nil
}

func GetCachePoolLayout() (map[string]*Reserved, error) {
	var return_err error
	cachePoolOnce.Do(func() {
		poolConf := NewCachePoolConfig()
		osConf := NewOSConfig()
		ways := GetCosInfo().CbmMaskLen

		if osConf.CacheWays+poolConf.Guarantee+poolConf.Besteffort+poolConf.Shared > uint(ways) {
			return_err = fmt.Errorf(
				"Error config: Guarantee + Besteffort + Shared + OS reserved ways should be less or equal to %d.", ways)
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

		if poolConf.Guarantee > 0 {
			wc := 1<<poolConf.Guarantee - 1
			resev, err := getReservedCache(wc,
				0,
				osConf.CacheWays,
				osCPUbm,
				syscaches)
			if err != nil {
				return_err = err
				return
			}
			cachePoolReserved[Guarantee] = resev
		}

		if poolConf.Besteffort > 0 {
			wc := 1<<poolConf.Besteffort - 1
			resev, err := getReservedCache(wc,
				poolConf.Guarantee,
				osConf.CacheWays,
				osCPUbm,
				syscaches)

			if err != nil {
				return_err = err
				return
			}
			cachePoolReserved[Besteffort] = resev
		}

		if poolConf.Shared > 0 {
			wc := 1<<poolConf.Shared - 1
			resev, err := getReservedCache(wc,
				poolConf.Guarantee+poolConf.Besteffort,
				osConf.CacheWays,
				osCPUbm,
				syscaches)

			if err != nil {
				return_err = err
				return
			}
			cachePoolReserved[Shared] = resev
			cachePoolReserved[Shared].Name = Shared
			cachePoolReserved[Shared].Quota = poolConf.MaxAllowedShared
		}
	})

	return cachePoolReserved, return_err

}
