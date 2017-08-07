package rdtpool

import (
	"sync"

	"openstackcore-rdtagent/util/bootcheck/infragroup"
	"openstackcore-rdtagent/util/bootcheck/osgroup"
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
//            use the mask in gurantee pool.
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
var ReservedInfo map[string]*Reserved
var revinfoOnce sync.Once

func GetReservedInfo() map[string]*Reserved {

	revinfoOnce.Do(func() {
		ReservedInfo = make(map[string]*Reserved, 10)

		r, err := osgroup.GetOSGroupReserve()
		if err == nil {
			ReservedInfo["os"] = &r
		}

		r, err = infragroup.GetInfraGroupReserve()
		if err == nil {
			ReservedInfo["infra"] = &r
		}
	})

	return ReservedInfo
}
