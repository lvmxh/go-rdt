package hospitality

// This model is just for cache info
// We can ref k8s

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"openstackcore-rdtagent/lib/cache"
	libcache "openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/cpu"
	_ "openstackcore-rdtagent/lib/proc"
	"openstackcore-rdtagent/lib/resctrl"
	libutil "openstackcore-rdtagent/lib/util"
	"openstackcore-rdtagent/model/policy"
)

/*
   Hospitality with details
*/

// consider RDT not only support the last level cache
type CacheScore map[string]map[string]uint32
type Hospitality struct {
	SC map[string]CacheScore `json:"score"`
}

func (h *Hospitality) getScoreByLevel(level uint32) error {

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

	// l2, l3
	cacheLevel := "l" + target_lev
	cacheS := make(map[string]map[string]uint32)
	h.SC = map[string]CacheScore{cacheLevel: cacheS}

	for _, sc := range syscaches {
		id, _ := strconv.Atoi(sc.Id)
		_, ok := h.SC[cacheLevel][sc.Id]
		if ok {
			// syscache.GetSysCaches returns caches per each CPU, there maybe
			// multiple cpus chares on same cache.
			continue
		} else {
			resaall := resctrl.GetResAssociation()
			ui32, _ := strconv.Atoi(sc.WaysOfAssociativity)
			numWays := uint32(ui32)

			var sb []*libutil.Bitmap
			for k, v := range resaall {
				if k == "infra" {
					continue
				}
				for _, sv := range v.Schemata {
					for _, cv := range sv {
						if cv.Id == uint8(id) {
							// FIXME we assume number of ways == length of cbm mask
							bm, _ := libutil.NewBitmap(int(numWays), cv.Mask)
							sb = append(sb, bm)
						}
					}
				}
			}

			inf := resctrl.GetRdtCosInfo()
			freeM := inf["l"+target_lev].CbmMask
			freeb, _ := libutil.NewBitmap(int(numWays), freeM)
			for _, v := range sb {
				freeb = freeb.Axor(v)
			}

			// avaliableWays = freeb.ToBinString()

			pf := cpu.GetMicroArch(cpu.GetSignature())
			if pf == "" {
				return fmt.Errorf("Unknow platform, please update the cpu_map.toml conf file.")
			}
			// FIXME add error check. This code is just for China Open days.
			p, _ := policy.GetPlatformPolicy(strings.ToLower(pf))
			ap := make(map[string]uint32)
			ap_counter := make(map[string]int)
			for _, pv := range p {
				for k, v := range pv {
					ap[k] = 0
					for _, cv := range v[0] {
						iv, err := strconv.Atoi(cv)
						if err != nil {
							return err
						}
						ap_counter[k] = iv
						break
					}
				}
			}
			fbs := freeb.ToBinStrings()
			for ak, av := range ap_counter {
				for _, v := range fbs {
					if v[0] == '1' {
						ap[ak] += uint32(len(v) / av)
					}
				}
				// To Percent
				// FIXME need to consider round
				ap[ak] = (ap[ak]*uint32(av)*100 + numWays/2) / numWays
			}

			cacheS[sc.Id] = ap
			h.SC[cacheLevel][sc.Id] = ap
		}
	}

	return nil
}

func (h *Hospitality) Get() error {
	level := libcache.GetLLC()
	return h.getScoreByLevel(level)
}
