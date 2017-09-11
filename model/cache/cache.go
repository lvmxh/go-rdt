package cache

// This model is just for cache info
// We can ref k8s

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	. "openstackcore-rdtagent/api/error"
	"openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/cpu"
	"openstackcore-rdtagent/lib/proc"
	"openstackcore-rdtagent/lib/resctrl"
	"openstackcore-rdtagent/model/policy"
	"openstackcore-rdtagent/util/rdtpool"
	"openstackcore-rdtagent/util/rdtpool/base"
)

// SizeMap is the map to bits of unit
var SizeMap = map[string]uint32{
	"K": 1024,
	"M": 1024 * 1024,
}

// CacheInfo with details
type CacheInfo struct {
	ID               uint32 `json:"cache_id"`
	NumWays          uint32
	NumSets          uint32
	NumPartitions    uint32
	LineSize         uint32
	TotalSize        uint32 `json:"total_size"`
	WaySize          uint32
	NumClasses       uint32
	WayContention    uint64
	CacheLevel       uint32
	Location         string            `json:"location_on_socket"`
	ShareCpuList     string            `json:"share_cpu_list"`
	AvaliableWays    string            `json:"avaliable_ways"`
	AvaliableCPUs    string            `json:"avaliable_cpus"`
	AvaliableIsoCPUs string            `json:"avaliable_isolated_cpus"`
	AvaliablePolicy  map[string]uint32 `json:"avaliable_policy"` // should move out here
}

type CacheInfos struct {
	Num    uint32               `json:"number"`
	Caches map[uint32]CacheInfo `json:"Caches"`
}

/*********************************************************************************/
type CacheSummary struct {
	Num int      `json:"number"`
	IDs []string `json:"caches_id"`
}

// Cat, Cqm seems CPU's feature.
// Should be better
// type Rdt struct {
// 	Cat   bool
// 	CatOn bool
// 	Cdp   bool
// 	CdpOn bool
// }
type CachesSummary struct {
	Rdt    bool                    `json:"rdt"`
	Cqm    bool                    `json:"cqm"`
	Cdp    bool                    `json:"cdp"`
	CdpOn  bool                    `json:"cdp_enable"`
	Cat    bool                    `json:"cat"`
	CatOn  bool                    `json:"cat_enable"`
	Caches map[string]CacheSummary `json:"caches"`
}

func (c *CachesSummary) getCaches() error {
	levs := syscache.AvailableCacheLevel()

	c.Caches = make(map[string]CacheSummary)
	for _, l := range levs {
		summary := &CacheSummary{}
		il, err := strconv.Atoi(l)
		if err != nil {
			return err
		}
		caches, err := syscache.GetSysCaches(il)
		if err != nil {
			return err
		}
		for _, v := range caches {
			summary.IDs = append(summary.IDs, v.Id)

		}
		summary.Num = len(caches)
		c.Caches["l"+l] = *summary
	}
	return nil
}

func (c *CachesSummary) Get() error {
	var err error
	var flag bool
	flag, err = proc.IsRdtAvailiable()
	if err != nil {
		return nil
	}
	c.Rdt = flag
	c.Cat = flag

	flag, err = proc.IsCqmAvailiable()
	if err != nil {
		return nil
	}
	c.Cqm = flag

	flag, err = proc.IsCdpAvailiable()
	if err != nil {
		return nil
	}
	c.Cdp = flag

	flag = proc.IsEnableCat()
	c.CatOn = flag

	flag = proc.IsEnableCdp()
	c.CdpOn = flag

	err = c.getCaches()
	if err != nil {
		return nil
	}

	return nil
}

// Convert a string cache size to uint32 in B
// eg: 1K = 1024
func ConvertCacheSize(size string) uint32 {
	unit := size[len(size)-1:]

	s := strings.TrimRight(size, unit)

	isize, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return uint32(isize) * SizeMap[unit]
}

func (c *CacheInfos) GetByLevel(level uint32) *AppError {

	llc := syscache.GetLLC()

	if llc != level {
		err := fmt.Errorf("Don't support cache level %d, Only expose last level cache %d", level, llc)
		return NewAppError(http.StatusBadRequest,
			"Error to get available cache", err)
	}

	target_lev := strconv.FormatUint(uint64(level), 10)

	// syscache.AvailableCacheLevel return []string
	levs := syscache.AvailableCacheLevel()
	sort.Strings(levs)

	syscaches, err := syscache.GetSysCaches(int(level))
	if err != nil {
		return NewAppError(http.StatusInternalServerError,
			"Error to get available cache", err)
	}

	cacheLevel := "L" + target_lev
	allres := resctrl.GetResAssociation()
	av, _ := rdtpool.GetAvailableCacheSchemata(allres, []string{"infra"}, "none", cacheLevel)

	c.Caches = make(map[uint32]CacheInfo)

	for _, sc := range syscaches {
		id, _ := strconv.Atoi(sc.Id)
		_, ok := c.Caches[uint32(id)]
		if ok {
			// syscache.GetSysCaches returns caches per each CPU, there maybe
			// multiple cpus chares on same cache.
			continue
		} else {
			// TODO: NumPartitions uint32,  NumClasses    uint32
			//       WayContention uint64,  Location string
			new_cacheinfo := CacheInfo{}

			ui32, _ := strconv.Atoi(sc.CoherencyLineSize)
			new_cacheinfo.LineSize = uint32(ui32)

			ui32, _ = strconv.Atoi(sc.NumberOfSets)
			new_cacheinfo.NumSets = uint32(ui32)

			// FIXME the relation between NumWays and sc.PhysicalLinePartition
			ui32, _ = strconv.Atoi(sc.WaysOfAssociativity)
			new_cacheinfo.NumWays = uint32(ui32)
			new_cacheinfo.WaySize = new_cacheinfo.LineSize * new_cacheinfo.NumSets

			new_cacheinfo.ID = uint32(id)
			new_cacheinfo.TotalSize = ConvertCacheSize(sc.Size)
			new_cacheinfo.ShareCpuList = sc.SharedCpuList
			new_cacheinfo.CacheLevel = level

			new_cacheinfo.AvaliableWays = av[sc.Id].ToBinString()

			cpuPools, _ := rdtpool.GetCPUPools()
			defaultCpus, _ := base.CpuBitmaps(resctrl.GetResAssociation()["."].CPUs)
			new_cacheinfo.AvaliableCPUs = cpuPools["all"][sc.Id].And(defaultCpus).ToBinString()
			new_cacheinfo.AvaliableIsoCPUs = cpuPools["isolated"][sc.Id].And(defaultCpus).ToBinString()

			pf := cpu.GetMicroArch(cpu.GetSignature())
			if pf == "" {
				return AppErrorf(http.StatusInternalServerError,
					"Unknow platform, please update the cpu_map.toml conf file")
			}
			// FIXME add error check. This code is just for China Open days.
			p, _ := policy.GetPlatformPolicy(strings.ToLower(pf))
			ap := make(map[string]uint32)
			//ap_counter := make(map[string]int)
			for _, pv := range p {
				// pv is policy.CATConfig.Catpolicy
				for t, _ := range pv {
					// t is the policy tier name
					tier, err := policy.GetPolicy(strings.ToLower(pf), t)
					if err != nil {
						return NewAppError(http.StatusInternalServerError,
							"Error to get policy", err)
					}

					iMax, err := strconv.Atoi(tier["MaxCache"])
					if err != nil {
						return NewAppError(http.StatusInternalServerError,
							"Error to get max cache", err)
					}
					iMin, err := strconv.Atoi(tier["MinCache"])
					if err != nil {
						return NewAppError(http.StatusInternalServerError,
							"Error to get min cache", err)
					}

					getAvailablePolicyCount(ap, iMax, iMin, allres, t, cacheLevel, sc.Id)

				}

			}
			new_cacheinfo.AvaliablePolicy = ap

			c.Caches[uint32(id)] = new_cacheinfo
			c.Num = c.Num + 1
		}
	}

	return nil
}

func getAvailablePolicyCount(ap map[string]uint32,
	iMax, iMin int,
	allres map[string]*resctrl.ResAssociation,
	tier, cacheLevel, cId string) error {

	var ways int

	reserved := rdtpool.GetReservedInfo()

	pool, _ := rdtpool.GetCachePoolName(uint32(iMax), uint32(iMin))

	switch pool {
	case rdtpool.Guarantee:
		ways = iMax
	case rdtpool.Besteffort:
		ways = iMin
	case rdtpool.Shared:
		// TODO get live count ?
		ap[tier] = uint32(reserved[rdtpool.Shared].Quota)
		return nil
	}

	pav, _ := rdtpool.GetAvailableCacheSchemata(allres, []string{"infra", "."}, pool, cacheLevel)
	ap[tier] = 0
	freeBitmapStrs := pav[cId].ToBinStrings()

	for _, val := range freeBitmapStrs {
		if val[0] == '1' {
			valLen := len(val)
			ap[tier] += uint32(valLen / ways)
		}
	}

	return nil
}
