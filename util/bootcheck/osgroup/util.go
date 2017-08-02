package osgroup

import (
	"fmt"
	"openstackcore-rdtagent/lib/cpu"
	. "openstackcore-rdtagent/lib/util"
)

// FIXME should find a good accommodation for CpuBitmaps
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
