package hospitality

// This model is just for cache info
// We can ref k8s

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	. "openstackcore-rdtagent/api/error"
	"openstackcore-rdtagent/db"
	"openstackcore-rdtagent/lib/cache"
	libcache "openstackcore-rdtagent/lib/cache"
	_ "openstackcore-rdtagent/lib/proc"
	"openstackcore-rdtagent/lib/proxyclient"
	libutil "openstackcore-rdtagent/lib/util"
	"openstackcore-rdtagent/model/policy"
	"openstackcore-rdtagent/util/rdtpool"
	"sort"
	"strconv"
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
			resaall := proxyclient.GetResAssociation()
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

			inf := proxyclient.GetRdtCosInfo()
			freeM := inf["l"+target_lev].CbmMask
			freeb, _ := libutil.NewBitmap(int(numWays), freeM)
			for _, v := range sb {
				freeb = freeb.Axor(v)
			}

			p, err := policy.GetDefaultPlatformPolicy()
			if err != nil {
				return err
			}

			ap := make(map[string]uint32)
			ap_counter := make(map[string]int)
			for _, pv := range p {
				// pv is policy.CATConfig.Catpolicy
				for k, _ := range pv {
					// k is the policy tier name
					ap[k] = 0
					tier, err := policy.GetDefaultPolicy(k)
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

func (h *HospitalityRaw) GetByRequest(req *HospitalityRequest) error {
	level := libcache.GetLLC()
	target_lev := strconv.FormatUint(uint64(level), 10)
	cacheLevel := "l" + target_lev
	cacheS := make(map[string]uint32)
	h.SC = map[string]CacheScoreRaw{cacheLevel: cacheS}

	max := req.MaxCache
	min := req.MinCache

	if req.Policy != "" {
		tier, err := policy.GetDefaultPolicy(req.Policy)
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

func (h *HospitalityRaw) GetByRequestMaxMin(max, min uint32, cache_id *uint32, target_lev string) error {

	var reqType string

	if max == 0 && min == 0 {
		reqType = rdtpool.Shared
	} else if max > min && min != 0 {
		reqType = rdtpool.Besteffort
	} else if max == min {
		reqType = rdtpool.Guarantee
	} else {
		return AppErrorf(http.StatusBadRequest,
			"Bad request, max_cache=%d, min_cache=%d", max, min)
	}

	resaall := proxyclient.GetResAssociation()

	av, _ := rdtpool.GetAvailableCacheSchemata(resaall, []string{"infra", "."}, reqType, "L"+target_lev)

	cacheS := make(map[string]uint32)
	h.SC = map[string]CacheScoreRaw{"l" + target_lev: cacheS}

	reserved := rdtpool.GetReservedInfo()

	if reqType == rdtpool.Shared {
		dbc, _ := db.NewDB()
		ws, _ := dbc.QueryWorkload(map[string]interface{}{
			"CosName": reserved[rdtpool.Shared].Name,
			"Status":  "Successful"})
		totalCount := reserved[rdtpool.Shared].Quota
		for k, _ := range av {
			if uint(len(ws)) < totalCount {
				cacheS[k] = 100
			} else {
				cacheS[k] = 0
			}
			retrimCache(k, cache_id, &cacheS)
		}
		return nil
	}

	for k, v := range av {
		var fbs []string
		cacheS[k] = 0

		fbs = v.ToBinStrings()

		log.Debugf("Free bitmask on cache [%s] is [%s]", k, fbs)
		// Calculate total supported
		for _, val := range fbs {
			if val[0] == '1' {
				valLen := len(val)
				if (valLen/int(min) > 0) && cacheS[k] < uint32(valLen) {
					cacheS[k] = uint32(valLen)
				}
			}
		}
		if cacheS[k] > 0 {
			// (NOTES): Gurantee will return 0|100
			// Besteffort will return (max continious ways) / max
			cacheS[k] = (cacheS[k] * 100) / max
			if cacheS[k] > 100 {
				cacheS[k] = 100
			}
		} else {
			cacheS[k] = 0
		}

		retrimCache(k, cache_id, &cacheS)
	}
	return nil
}

func retrimCache(cacheId string, cache_id *uint32, cacheS *map[string]uint32) {

	icacheId, _ := strconv.Atoi(cacheId)
	if cache_id != nil {
		// We only care about specific cache_id
		if *cache_id != uint32(icacheId) {
			delete(*cacheS, cacheId)
		}
	}
}
