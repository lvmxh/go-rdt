// +build linux

package syscache

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

const (
	SysCpuPath = "/sys/devices/system/cpu/"
)

// All the type a string.
type SysCache struct {
	CoherencyLineSize     string
	Id                    string
	Level                 string
	NumberOfSets          string
	PhysicalLinePartition string
	SharedCpuList         string
	SharedCpuMap          string
	Size                  string
	Type                  string
	WaysOfAssociativity   string
	// Power              string
	// Uevent             string
}

// Only support string At present.
func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	val := reflect.ValueOf(value)

	structFieldValue.Set(val)
	return nil
}

// /sys/devices/system/cpu/cpu*/cache/index*/*
// pass var caches map[string]SysCache
/*
usage:
    ignore := []string{"uevent"}
    syscache := &SysCache{}
	filepath.Walk(dir, getSysCache(ignore, syscache))
*/
func getSysCache(ignore []string, cache *SysCache) filepath.WalkFunc {

	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// add log
			return nil
		}

		// ignore dir.
		f := filepath.Base(path)
		if info.IsDir() {
			for _, d := range ignore {
				if d == f {
					return filepath.SkipDir
				}
			}
			return nil
		}
		for _, d := range ignore {
			if d == f {
				return nil
			}
		}

		name := strings.Replace(strings.Title(strings.Replace(f, "_", " ", -1)), " ", "", -1)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			// add log
			return err
		}
		return SetField(cache, name, strings.TrimSpace(string(data)))
	}
}

// Traverse all sys cache file for a specify level
func GetSysCaches(level int) (map[string]SysCache, error) {
	ignore := []string{"uevent", "power"}
	caches := make(map[string]SysCache)
	files, err := filepath.Glob(SysCpuPath + "cpu*/cache/index" + strconv.Itoa(level))
	if err != nil {
		return caches, err
	}

	for _, f := range files {
		cache := &SysCache{}
		err := filepath.Walk(f, getSysCache(ignore, cache))
		if err != nil {
			return caches, err
		}
		if _, ok := caches[cache.Id]; !ok {
			caches[cache.Id] = *cache
		}
	}
	return caches, nil
}

// Just return the L2 and L3 level cache, strip L1 cache.
// By default, get the info from cpu0 path, any issue?
// The type of return should be string or int?
func AvailableCacheLevel() []string {
	var levels []string
	files, _ := filepath.Glob(SysCpuPath + "cpu0/cache/index*/level")
	for _, f := range files {
		dat, _ := ioutil.ReadFile(f)
		sdat := strings.TrimRight(string(dat), "\n")
		if 0 != strings.Compare("1", sdat) {
			levels = append(levels, sdat)
		}
	}
	return levels
}
