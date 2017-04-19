package model

// This model is just for cache info
// We can ref k8s

import (
	"strconv"

	"openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/proc"
)

type CacheDetail struct {
	ID           string `json:"id"`
	Location     string `json:"location_on_socket"`
	ShareCpuList string `json:"share_cpu_list"`
}

// Are cache symmetric on different socket?
type CacheInfo struct {
	NumWays       uint32
	NumSets       uint32
	NumPartitions uint32
	LineSize      uint32
	TotalSize     uint32 `json:"total_size"`
	WaySize       uint32
	NumClasses    uint32
	WayContention uint64
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
