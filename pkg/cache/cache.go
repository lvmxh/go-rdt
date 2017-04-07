package cache

import (
	"sync"
)

// cpu map list
type CpuMap []uint32

// cache bank
type Cache struct {
	Id      uint32
	Level   uint32
	CpuList CpuMap
	Cdp     bool
	CdpOn   bool
}

type CacheInfo struct {
	Num    uint32
	Caches []*Cache
}

var once sync.Once
var lock sync.Mutex
var l3cacheinfo *CacheInfo
var l2cacheinfo *CacheInfo

func Initialize(l3 CacheInfo, l2 CacheInfo) {
	once.Do(func() {
		l3cacheinfo = &l3
		l2cacheinfo = &l2
	})
}

func GetL3CacheInfo() CacheInfo {
	lock.Lock()
	defer lock.Unlock()
	if l3cacheinfo == nil {
		l3cacheinfo = &CacheInfo{}
	}
	return *l3cacheinfo
}

func GetL2CacheInfo() CacheInfo {
	lock.Lock()
	defer lock.Unlock()
	if l2cacheinfo == nil {
		l2cacheinfo = &CacheInfo{}
	}
	return *l2cacheinfo
}
