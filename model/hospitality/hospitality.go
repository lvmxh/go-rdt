package hospitality

// This model is just for cache info
// We can ref k8s

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	. "openstackcore-rdtagent/api/error"
	"openstackcore-rdtagent/lib/cache"
	libcache "openstackcore-rdtagent/lib/cache"
	"openstackcore-rdtagent/lib/cpu"
	_ "openstackcore-rdtagent/lib/proc"
	"openstackcore-rdtagent/lib/resctrl"
	libutil "openstackcore-rdtagent/lib/util"
	"openstackcore-rdtagent/model/policy"
	"openstackcore-rdtagent/util/rdtpool"
	. "openstackcore-rdtagent/util/rdtpool/base"
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

	var reqType string

	if max == 0 {
		reqType = rdtpool.Shared
	} else if max > min && min != 0 {
		reqType = rdtpool.Besteffort
	} else if max == min {
		reqType = rdtpool.Gurantee
	} else {
		return AppErrorf(http.StatusBadRequest,
			"Bad request, max_cache=%d, min_cache=%d", max, min)
	}

	// TODO: Need to calculate how many for this kinds workload already
	// running. For now treat it as max = 1 to avoid devided by zero. The
	// score is not accruate for max = min = 0
	if max == 0 {
		max = 1
	}

	resaall := resctrl.GetResAssociation()

	av, _ := rdtpool.GetAvailableCacheSchemata(resaall, []string{"infra", "."}, reqType, "L"+target_lev)

	cacheS := make(map[string]uint32)
	h.SC = map[string]CacheScoreRaw{"l" + target_lev: cacheS}

	numWays := uint32(GetCosInfo().CbmMaskLen)

	if reqType == rdtpool.Besteffort {
		reserved := rdtpool.GetReservedInfo()

		for k, v := range av {
			var totalCount = 0
			cacheS[k] = 0
			cacheId, _ := strconv.Atoi(k)
			sharedBm := reserved[rdtpool.Shared].Schemata[k]
			besteffortBm := reserved[rdtpool.Besteffort].Schemata[k]
			// Please read it and to understand it
			// Hard to describe it in human English.
			if besteffortBm.Axor(sharedBm).IsEmpty() ||
				sharedBm.GetConnectiveBits(max-min, 0, true).IsEmpty() {
				log.Infof("No cache way left in besteffort pool on cache id %s", k)
				continue
			}

			fbs := besteffortBm.Axor(sharedBm).ToBinStrings()
			// Calculate total supported
			for _, val := range fbs {
				if val[0] == '1' {
					totalCount += len(val) / int(min)
				}
			}
			if totalCount == 0 {
				continue
			}

			log.Debugf("Free overlap bitmask on cache [%s] is [%s]", k, v.ToBinStrings())
			fbs = v.Axor(sharedBm).ToBinStrings()
			// Scan no-overlap ways
			log.Debugf("Free executive bitmask on cache [%s] is [%s]", k, fbs)
			for _, val := range fbs {
				if val[0] == '1' {
					cacheS[k] += uint32(len(val) / int(min))
				}
			}

			cacheS[k] = cacheS[k] * 100 / uint32(totalCount)
			if cache_id != nil {
				// We only care about specific cache_id
				if *cache_id != uint32(cacheId) {
					delete(cacheS, k)
				}
			}
		}
		return nil
	}

	// Notes that av is a map, so we are not sure about the cache id order
	for k, v := range av {
		log.Debugf("Free bitmask on cache [%s] is [%s]", k, v.ToBinString())
		fbs := v.ToBinStrings()
		cacheS[k] = 0
		cacheId, _ := strconv.Atoi(k)

		for _, val := range fbs {
			if val[0] == '1' {
				cacheS[k] += uint32(len(val) / int(max))
			}
		}
		// Conver to percentage
		cacheS[k] = (cacheS[k]*uint32(max)*100 + numWays/2) / numWays

		if cache_id != nil {
			// We only care about specific cache_id
			if *cache_id != uint32(cacheId) {
				delete(cacheS, k)
			}
		}
	}

	return nil
}
