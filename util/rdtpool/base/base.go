package base

import (
	"fmt"
	"strconv"
	"sync"

	"openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/cpu"
	"openstackcore-rdtagent/lib/resctrl"
	. "openstackcore-rdtagent/lib/util"
)

// FIXME should find a good accommodation for file

type CosInfo struct {
	CbmMaskLen int
	MinCbmBits int
	NumClosids int
}

var catCosInfo = &CosInfo{0, 0, 0}
var infoOnce sync.Once

// Concurrency-safe.
func GetCosInfo() CosInfo {
	infoOnce.Do(func() {
		rcinfo := resctrl.GetRdtCosInfo()
		level := syscache.GetLLC()
		target_lev := strconv.FormatUint(uint64(level), 10)
		cacheLevel := "l" + target_lev

		catCosInfo.CbmMaskLen = CbmLen(rcinfo[cacheLevel].CbmMask)
		catCosInfo.MinCbmBits = rcinfo[cacheLevel].MinCbmBits
		catCosInfo.NumClosids = rcinfo[cacheLevel].NumClosids
	})
	return *catCosInfo
}

// a wraper for Bitmap
func CpuBitmaps(cpuids interface{}) (*Bitmap, error) {
	// FIXME need a wrap for CPU bitmap.
	cpunum := cpu.HostCpuNum()
	if cpunum == 0 {
		// return nil or an empty Bitmap?
		var bm *Bitmap
		return bm, fmt.Errorf("Unable to get Total CPU numbers on Host")
	}
	return NewBitmap(cpunum, cpuids)
}

func CacheBitmaps(bitmask interface{}) (*Bitmap, error) {
	// FIXME need a wrap for CPU bitmap.
	len := GetCosInfo().CbmMaskLen
	if len == 0 {
		// return nil or an empty Bitmap?
		var bm *Bitmap
		return bm, fmt.Errorf("Unable to get Total cache ways on Host")
	}
	return NewBitmap(len, bitmask)
}