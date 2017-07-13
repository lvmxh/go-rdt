package workload

import (
	"reflect"
	"testing"

	"openstackcore-rdtagent/lib/resctrl"
	"openstackcore-rdtagent/model/cache"
)

func testGetGroupNames(w *RDTWorkLoad, m map[string]*resctrl.ResAssociation, b, n string, s []string, t *testing.T) {
	base_grp, new_grp, sub_grps := getGroupNames(w, m)
	if b != base_grp {
		t.Errorf("wrong base group!")
	}
	if n != new_grp {
		t.Errorf("wrong new group!")
	}
	if !reflect.DeepEqual(sub_grps, s) {
		t.Errorf("wrong sub group list!")
	}
}

func TestGetGroupNamesEmptyGroupName(t *testing.T) {
	w := RDTWorkLoad{TaskIDs: []string{"123"}}
	base_grp := "."
	new_grp := "123"
	sub_grps := []string{}

	m := map[string]*resctrl.ResAssociation{
		".": &resctrl.ResAssociation{},
		"1": &resctrl.ResAssociation{},
	}

	testGetGroupNames(&w, m, base_grp, new_grp, sub_grps, t)
}

func TestGetGroupNamesWithGroupName(t *testing.T) {
	w := RDTWorkLoad{Group: []string{"abc"}}
	base_grp := "abc"
	new_grp := "abc"
	sub_grps := []string{}

	m := map[string]*resctrl.ResAssociation{
		".":   &resctrl.ResAssociation{},
		"abc": &resctrl.ResAssociation{},
	}

	testGetGroupNames(&w, m, base_grp, new_grp, sub_grps, t)
}

func TestGetGroupNamesWithTwoGroupName(t *testing.T) {
	w := RDTWorkLoad{Group: []string{"abc", "def"}}
	base_grp := "abc"
	new_grp := "def"
	sub_grps := []string{}

	m := map[string]*resctrl.ResAssociation{
		".":   &resctrl.ResAssociation{},
		"abc": &resctrl.ResAssociation{},
	}

	testGetGroupNames(&w, m, base_grp, new_grp, sub_grps, t)
}

func TestGetGroupNamesWithGroupNameAndSub(t *testing.T) {
	w := RDTWorkLoad{Group: []string{"abc", "def"}}
	base_grp := "abc"
	new_grp := "def"
	sub_grps := []string{"abc-sub"}

	m := map[string]*resctrl.ResAssociation{
		".":       &resctrl.ResAssociation{},
		"abc":     &resctrl.ResAssociation{},
		"abc-sub": &resctrl.ResAssociation{},
	}

	testGetGroupNames(&w, m, base_grp, new_grp, sub_grps, t)
}

func TestCalculateOffset(t *testing.T) {
	cos1 := resctrl.CacheCos{0, "f"}
	cos2 := resctrl.CacheCos{0, "10"}

	r := map[string]*resctrl.ResAssociation{
		"sub1": &resctrl.ResAssociation{Schemata: map[string][]resctrl.CacheCos{"L3": []resctrl.CacheCos{cos1}}},
		"sub2": &resctrl.ResAssociation{Schemata: map[string][]resctrl.CacheCos{"L3": []resctrl.CacheCos{cos2}}},
	}

	sub_grp := []string{"sub2", "sub1"}
	if (calculateOffset(r, sub_grp, "L3", 0)) != 5 {
		t.Errorf("wrong offset")
	}
}

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
