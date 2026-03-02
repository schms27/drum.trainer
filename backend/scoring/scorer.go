package scoring

import (
	"math"
	"sort"

	"github.com/simu/drum-trainer/backend/models"
)

const matchWindowMs = 200.0

// Score computes a ScoreResult by aligning recordedHits to targetNotes.
//
// Algorithm: greedy nearest-neighbour with instrument grouping.
//   - Matching is performed per instrument group so a bass drum hit cannot
//     steal a snare match.
//   - Within each group, target notes and recorded hits are time-sorted, then
//     a linear scan finds the nearest unmatched recorded hit within ±200ms.
//   - Unmatched target notes → misses.
//   - Unmatched recorded hits → ghost notes.
func Score(targetNotes []models.NoteEvent, recordedHits []models.RecordedHit, instrumentGroup func(int) string) *models.ScoreResult {
	// Group targets by instrument.
	targetsByGroup := make(map[string][]models.NoteEvent)
	for _, n := range targetNotes {
		g := instrumentGroup(n.NoteNumber)
		targetsByGroup[g] = append(targetsByGroup[g], n)
	}
	// Group recorded hits by instrument.
	hitsByGroup := make(map[string][]models.RecordedHit)
	for _, h := range recordedHits {
		g := instrumentGroup(h.NoteNumber)
		hitsByGroup[g] = append(hitsByGroup[g], h)
	}

	// Collect all groups (union of target and recorded groups).
	groupSet := make(map[string]struct{})
	for g := range targetsByGroup {
		groupSet[g] = struct{}{}
	}
	for g := range hitsByGroup {
		groupSet[g] = struct{}{}
	}
	groups := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groups = append(groups, g)
	}
	sort.Strings(groups)

	var (
		allMatchedHits []models.MatchedHit
		allGhostHits   []models.GhostHit
		allOffsets     []float64
		totalTargets   int
		totalMatched   int
		totalMissed    int
		totalGhosts    int
		totalRecorded  int
		byInstrument   []models.InstrumentBreakdown
	)

	for _, group := range groups {
		targets := targetsByGroup[group]
		hits := hitsByGroup[group]

		// Sort both slices by time.
		sort.Slice(targets, func(i, j int) bool { return targets[i].TimeMs < targets[j].TimeMs })
		sort.Slice(hits, func(i, j int) bool { return hits[i].TimeMs < hits[j].TimeMs })

		usedHit := make([]bool, len(hits))
		var groupOffsets []float64
		var groupMatched []models.MatchedHit

		lowerBound := 0 // advance this to skip hits that are too early

		for _, target := range targets {
			// Move lower bound past hits that are definitely too early.
			for lowerBound < len(hits) && hits[lowerBound].TimeMs < target.TimeMs-matchWindowMs {
				lowerBound++
			}

			bestIdx := -1
			bestDist := math.MaxFloat64

			for k := lowerBound; k < len(hits); k++ {
				delta := hits[k].TimeMs - target.TimeMs
				if delta > matchWindowMs {
					break // all subsequent hits are too late
				}
				if usedHit[k] {
					continue
				}
				dist := math.Abs(delta)
				if dist < bestDist {
					bestDist = dist
					bestIdx = k
				}
			}

			if bestIdx >= 0 {
				usedHit[bestIdx] = true
				offset := hits[bestIdx].TimeMs - target.TimeMs
				groupOffsets = append(groupOffsets, offset)
				allOffsets = append(allOffsets, offset)
				mh := models.MatchedHit{
					TargetTimeMs:   target.TimeMs,
					RecordedTimeMs: hits[bestIdx].TimeMs,
					OffsetMs:       offset,
					NoteNumber:     target.NoteNumber,
					Instrument:     group,
				}
				groupMatched = append(groupMatched, mh)
				allMatchedHits = append(allMatchedHits, mh)
				totalMatched++
			} else {
				totalMissed++
			}
		}

		// Collect ghost hits (unmatched recorded hits).
		groupGhosts := 0
		for k, used := range usedHit {
			if !used {
				groupGhosts++
				allGhostHits = append(allGhostHits, models.GhostHit{
					RecordedTimeMs: hits[k].TimeMs,
					NoteNumber:     hits[k].NoteNumber,
					Instrument:     group,
				})
			}
		}
		totalGhosts += groupGhosts
		totalTargets += len(targets)
		totalRecorded += len(hits)

		// Per-instrument breakdown.
		bd := models.InstrumentBreakdown{
			Instrument:   group,
			TargetCount:  len(targets),
			MatchedCount: len(groupMatched),
			MissedCount:  len(targets) - len(groupMatched),
			GhostCount:   groupGhosts,
			HitRate:      safeDiv(float64(len(groupMatched)), float64(len(targets))),
		}
		if len(groupOffsets) > 0 {
			avg := mean(groupOffsets)
			std := stddev(groupOffsets, avg)
			bd.AvgOffsetMs = &avg
			bd.StdOffsetMs = &std
		}
		byInstrument = append(byInstrument, bd)
	}

	globalAvg := 0.0
	globalStd := 0.0
	if len(allOffsets) > 0 {
		globalAvg = mean(allOffsets)
		globalStd = stddev(allOffsets, globalAvg)
	}

	return &models.ScoreResult{
		HitRate:      safeDiv(float64(totalMatched), float64(totalTargets)),
		MissRate:     safeDiv(float64(totalMissed), float64(totalTargets)),
		GhostRate:    safeDiv(float64(totalGhosts), float64(totalRecorded)),
		AvgOffsetMs:  globalAvg,
		StdOffsetMs:  globalStd,
		ByInstrument: byInstrument,
		MatchedHits:  allMatchedHits,
		GhostHits:    allGhostHits,
	}
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func stddev(vals []float64, avg float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	variance := 0.0
	for _, v := range vals {
		d := v - avg
		variance += d * d
	}
	return math.Sqrt(variance / float64(len(vals)))
}
