package rdtpool

import (
	"fmt"
	"strconv"
	"sync"

	"openstackcore-rdtagent/lib/resctrl"
	libutil "openstackcore-rdtagent/lib/util"
	"openstackcore-rdtagent/util"
	. "openstackcore-rdtagent/util/rdtpool/base"
)

// A map that contains all reserved resource information
// Use resource key as index, the key could be as following:
//
// OS: os group information, it contains reserved cache information.
//     Required.
// INFRA: infra group information.
//        Optional.
// GUARANTEE: guarantee cache pool information, it's a pool instead of a
//            resource group. When try to allocate max_cache = min_cache,
//            use the mask in guarantee pool.
//            Optional.
// BESTEFFORT: besteffort pool information, it's a pool instead of a resource
//             group. When try to allocate max_cache > min_cache, allocate
//             from this pool.
//             Optional.
// SHARED: shared group, it's a resource group instead of a pool. When try
//         to allocate max_cache == min_cache == 0, just add cpus, tasks IDs
//         to this resource group. Need to count how many workload in this
//         resource group when calculating hosptility score.
//         Optional

const (
	// Resource group
	OS    = "os"
	Infra = "infra"
	// Cache resource pool
	Guarantee   = "guarantee"
	Besteffort = "besteffort"
	Shared     = "shared"
)

var ReservedInfo map[string]*Reserved
var revinfoOnce sync.Once

func GetReservedInfo() map[string]*Reserved {

	revinfoOnce.Do(func() {
		ReservedInfo = make(map[string]*Reserved, 10)

		r, err := GetOSGroupReserve()
		if err == nil {
			ReservedInfo[OS] = &r
		}

		r, err = GetInfraGroupReserve()
		if err == nil {
			ReservedInfo[Infra] = &r
		}

		poolinfo, err := GetCachePoolLayout()
		if err == nil {
			for k, v := range poolinfo {
				ReservedInfo[k] = v
			}
		}
	})

	return ReservedInfo
}

// Return available schemata of caches from specific pool: guarantee,
// besteffort, shared or just none
func GetAvailableCacheSchemata(allres map[string]*resctrl.ResAssociation,
	ignore_groups []string,
	pool string,
	cacheLevel string) (map[string]*libutil.Bitmap, error) {

	GetReservedInfo()
	// FIXME (Shaohe) A central util to generate schemata Bitmap
	schemata := map[string]*libutil.Bitmap{}

	if pool == "none" {
		for k, _ := range ReservedInfo[OS].Schemata {
			schemata[k], _ = CacheBitmaps(GetCosInfo().CbmMask)
		}
	} else {
		resv, ok := ReservedInfo[pool]
		if !ok {
			return nil, fmt.Errorf("error doesn't support pool %s", pool)
		}

		for k, v := range resv.Schemata {
			schemata[k] = v
		}
	}

	for k, v := range allres {
		if util.HasElem(ignore_groups, k) {
			continue
		}
		if sv, ok := v.Schemata[cacheLevel]; ok {
			for _, cv := range sv {
				k := strconv.Itoa(int(cv.Id))
				bm, _ := CacheBitmaps(cv.Mask)
				// And check cpu list is empty
				if cv.Mask == GetCosInfo().CbmMask {
					continue
				}
				schemata[k] = schemata[k].Axor(bm)
			}
		}
	}
	return schemata, nil
}
