package model

// This model is just for cache info
// We can ref k8s

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/proc"
)

var SizeMap = map[string]uint32{
	"K": 1024,
	"M": 1024 * 1024,
}

/*
   CacheInfo with details
*/
type CacheInfo struct {
	ID            uint32 `json:"cache_id"`
	NumWays       uint32
	NumSets       uint32
	NumPartitions uint32
	LineSize      uint32
	TotalSize     uint32 `json:"total_size"`
	WaySize       uint32
	NumClasses    uint32
	WayContention uint64
	CacheLevel    uint32
	Location      string `json:"location_on_socket"`
	ShareCpuList  string `json:"share_cpu_list"`
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

func (c *CacheInfos) GetByLevel(level uint32) error {

	target_lev := strconv.FormatUint(uint64(level), 10)

	// syscache.AvailableCacheLevel return []string
	levs := syscache.AvailableCacheLevel()
	sort.Strings(levs)
	i := sort.SearchStrings(levs, target_lev)
	if i < len(levs) && levs[i] == target_lev {

	} else {
		err := fmt.Errorf("Could not found cache level %s on host", target_lev)
		return err
	}

	syscaches, err := syscache.GetSysCaches(int(level))
	if err != nil {
		return err
	}

	c.Caches = make(map[uint32]CacheInfo)

	for _, sc := range syscaches {
		id, _ := strconv.Atoi(sc.Id)
		_, ok := c.Caches[uint32(id)]
		if ok {
			// syscache.GetSysCaches returns caches per each CPU, there maybe
			// multiple cpus chares on same cache.
			continue
		} else {
			new_cacheinfo := CacheInfo{}
			new_cacheinfo.ID = uint32(id)
			new_cacheinfo.TotalSize = ConvertCacheSize(sc.Size)
			new_cacheinfo.ShareCpuList = sc.SharedCpuList
			new_cacheinfo.CacheLevel = level
			c.Caches[uint32(id)] = new_cacheinfo
			c.Num = c.Num + 1
		}
	}

	return nil
}
