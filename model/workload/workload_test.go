package workload

import (
	"testing"

	"openstackcore-rdtagent/model/cache"
)

func TestGetCacheIDs(t *testing.T) {
	cacheinfos := &cache.CacheInfos{Num: 2,
		Caches: map[uint32]cache.CacheInfo{
			0: cache.CacheInfo{ID: 0, ShareCpuList: "0-3"},
			1: cache.CacheInfo{ID: 1, ShareCpuList: "4-7"},
		}}

	cpubitmap := "3"

	cache_ids := getCacheIDs(cpubitmap, cacheinfos, 8)
	if len(cache_ids) != 1 && cache_ids[0] != 0 {
		t.Errorf("cache_ids should be [0], but we get %v", cache_ids)
	}

	cpubitmap = "1f"
	cache_ids = getCacheIDs(cpubitmap, cacheinfos, 8)
	if len(cache_ids) != 2 {
		t.Errorf("cache_ids should be [0, 1], but we get %v", cache_ids)
	}

	cpubitmap = "10"
	cache_ids = getCacheIDs(cpubitmap, cacheinfos, 8)
	if len(cache_ids) != 1 && cache_ids[0] != 1 {
		t.Errorf("cache_ids should be [1], but we get %v", cache_ids)
	}

	cpubitmap = "f00"
	cache_ids = getCacheIDs(cpubitmap, cacheinfos, 8)
	if len(cache_ids) != 0 {
		t.Errorf("cache_ids should be [], but we get %v", cache_ids)
	}

}
