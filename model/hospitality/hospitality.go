package hospitality

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
	libcache "openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/cpu"
	_ "openstackcore-rdtagent/lib/proc"
	"openstackcore-rdtagent/lib/resctrl"
	libutil "openstackcore-rdtagent/lib/util"
	"openstackcore-rdtagent/model/policy"
	modelutil "openstackcore-rdtagent/model/util"
	"openstackcore-rdtagent/model/workload"
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
				// pv is policy.CATConfig.Catpolicy
				for k, _ := range pv {
					// k is the policy tier name
					ap[k] = 0
					tier, err := policy.GetPolicy(strings.ToLower(pf), k)
					if err != nil {
						return err
					}
					iv, err := strconv.Atoi(tier["MaxCache"])
					if err != nil {
						return err
					}
					ap_counter[k] = iv
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
		}
	}

	return nil
}

func (h *Hospitality) Get() error {
	level := libcache.GetLLC()
	return h.getScoreByLevel(level)
}

///////////////////////////////////////////////////////////////
//  Support to give hospitality score by request             //
///////////////////////////////////////////////////////////////
// Hospitality score request
type HospitalityRequest struct {
	MaxCache uint32  `json:"max_cache,omitempty"`
	MinCache uint32  `json:"min_cache,omitempty"`
	Policy   string  `json:"policy,omitempty"`
	CacheId  *uint32 `json:"cache_id,omitempty"`
}

/*
{
	"score": {
		"l3": {
			"0": 30
			"1": 30
		}
	}
}
*/
type CacheScoreRaw map[string]uint32
type HospitalityRaw struct {
	SC map[string]CacheScoreRaw `json:"score"`
}

func (h *HospitalityRaw) GetByRequest(req *HospitalityRequest) *AppError {
	level := libcache.GetLLC()
	target_lev := strconv.FormatUint(uint64(level), 10)
	cacheLevel := "l" + target_lev
	cacheS := make(map[string]uint32)
	h.SC = map[string]CacheScoreRaw{cacheLevel: cacheS}

	max := req.MaxCache
	min := req.MinCache

	if req.Policy != "" {
		pf := cpu.GetMicroArch(cpu.GetSignature())
		if pf == "" {
			return AppErrorf(http.StatusInternalServerError,
				"Unknow platform, please update the cpu_map.toml conf file.")
		}
		tier, err := policy.GetPolicy(strings.ToLower(pf), req.Policy)
		if err != nil {
			return NewAppError(http.StatusInternalServerError,
				"Can not find Policy", err)
		}
		m, _ := strconv.Atoi(tier["MaxCache"])
		n, _ := strconv.Atoi(tier["MinCache"])
		max = uint32(m)
		min = uint32(n)
	}
	return h.GetByRequestMaxMin(max, min, req.CacheId, target_lev)
}

func (h *HospitalityRaw) GetByRequestMaxMin(max, min uint32, cache_id *uint32, target_lev string) *AppError {

	// TODO: for max > min > 0 we need to wait for besteffort pool get implemented.
	if max != min {
		err := fmt.Errorf("Don't support max != mix case yet!")
		return NewAppError(http.StatusBadRequest, "Bad request", err)
	}

	cacheS := make(map[string]uint32)
	h.SC = map[string]CacheScoreRaw{"l" + target_lev: cacheS}

	// TODO: Need to calculate how many workload for this kinds already
	// running. For now treat it as max = min = 1
	if max == min && max == 0 {
		max = 1
		min = 1
	}

	resaall := resctrl.GetResAssociation()
	rdtinfo := resctrl.GetRdtCosInfo()

	// ignore "." and "infra" group for now
	grp := workload.CalculateDefaultGroup(resaall, []string{".", "infra"}, false)
	catinfo, ok := rdtinfo[strings.ToLower("l"+target_lev)]

	numWays := uint32(modelutil.CbmLen(catinfo.CbmMask))

	if !ok {
		err := fmt.Errorf("Don't support cache level l%s", target_lev)
		return NewAppError(http.StatusBadRequest, "Bad request", err)
	}

	for _, schemata := range grp.Schemata["L"+target_lev] {
		id := strconv.FormatUint(uint64(schemata.Id), 10)
		freeb, _ := libutil.NewBitmap(schemata.Mask)
		fbs := freeb.ToBinStrings()

		cacheS[id] = 0
		for _, v := range fbs {
			if v[0] == '1' {
				cacheS[id] += uint32(len(v) / int(max))
			}
		}
		// Conver to percentage
		cacheS[id] = (cacheS[id]*uint32(max)*100 + numWays/2) / numWays

		if cache_id != nil {
			// We only care about specific cache_id
			if *cache_id != uint32(schemata.Id) {
				delete(cacheS, id)
			}
		}
	}
	return nil
}
