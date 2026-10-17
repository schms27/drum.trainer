package scoring

import (
	"math"
	"sort"

	"github.com/simu/drum-trainer/backend/models"
)

const (
	// maxMatchWindowMs is the widest a target's match window can be: an
	// isolated note accepts hits up to ±200ms away.
	maxMatchWindowMs = 200.0
	// minMatchWindowMs is the floor for densely packed notes (flams, fast
	// rolls) so they still have a usable window.
	minMatchWindowMs = 25.0
)

// Score computes a ScoreResult by aligning recordedHits to targetNotes.
//
// Algorithm:
//   - Matching is performed per instrument group so a bass drum hit cannot
//     steal a snare match.
//   - Each target gets its own window: half the distance to its nearest
//     neighbouring target in the same group, clamped to [25ms, 200ms]. In fast
//     passages windows therefore stop at the midpoint between notes, so a hit
//     is judged against the note it was closest to rather than a neighbour.
//   - All (target, hit) pairs inside a window are assigned closest-first, so
//     an earlier target cannot take a hit that is much closer to a later one.
//   - Unmatched target notes → misses.
//   - Unmatched recorded hits → ghost notes.
func Score(targetNotes []models.NoteEvent, recordedHits []models.RecordedHit, instrumentGroup func(int) string) *models.ScoreResult {
	return ScoreWithGroups(targetNotes, recordedHits, instrumentGroup, instrumentGroup)
}

// ScoreWithGroups is Score with separate note→instrument mappings for the
// target notes (from the MIDI file) and the recorded hits (from the kit),
// since a file and a kit may number the same drum differently.
func ScoreWithGroups(targetNotes []models.NoteEvent, recordedHits []models.RecordedHit, targetGroup, hitGroup func(int) string) *models.ScoreResult {
	// Group targets by instrument.
	targetsByGroup := make(map[string][]models.NoteEvent)
	for _, n := range targetNotes {
		g := targetGroup(n.NoteNumber)
		targetsByGroup[g] = append(targetsByGroup[g], n)
	}
	// Group recorded hits by instrument.
	hitsByGroup := make(map[string][]models.RecordedHit)
	for _, h := range recordedHits {
		g := hitGroup(h.NoteNumber)
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

		matchOf, usedHit := matchGroup(targets, hits)
		var groupOffsets []float64
		var groupMatched []models.MatchedHit

		for i, target := range targets {
			k := matchOf[i]
			if k < 0 {
				totalMissed++
				continue
			}
			offset := hits[k].TimeMs - target.TimeMs
			groupOffsets = append(groupOffsets, offset)
			allOffsets = append(allOffsets, offset)
			mh := models.MatchedHit{
				TargetTimeMs:   target.TimeMs,
				RecordedTimeMs: hits[k].TimeMs,
				OffsetMs:       offset,
				NoteNumber:     target.NoteNumber,
				Instrument:     group,
			}
			groupMatched = append(groupMatched, mh)
			allMatchedHits = append(allMatchedHits, mh)
			totalMatched++
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

// matchGroup pairs time-sorted targets with time-sorted hits of one instrument
// group. It returns, for each target, the index of its matched hit (or -1),
// and for each hit whether it was used.
func matchGroup(targets []models.NoteEvent, hits []models.RecordedHit) (matchOf []int, usedHit []bool) {
	type candidate struct {
		target, hit int
		dist        float64
	}
	var candidates []candidate
	for i, t := range targets {
		w := matchWindow(targets, i)
		// First hit not earlier than the window start.
		k := sort.Search(len(hits), func(k int) bool { return hits[k].TimeMs >= t.TimeMs-w })
		for ; k < len(hits) && hits[k].TimeMs <= t.TimeMs+w; k++ {
			candidates = append(candidates, candidate{i, k, math.Abs(hits[k].TimeMs - t.TimeMs)})
		}
	}
	// Closest pairs first; the stable sort keeps ties in time order.
	sort.SliceStable(candidates, func(a, b int) bool { return candidates[a].dist < candidates[b].dist })

	matchOf = make([]int, len(targets))
	for i := range matchOf {
		matchOf[i] = -1
	}
	usedHit = make([]bool, len(hits))
	for _, c := range candidates {
		if matchOf[c.target] < 0 && !usedHit[c.hit] {
			matchOf[c.target] = c.hit
			usedHit[c.hit] = true
		}
	}
	return matchOf, usedHit
}

// matchWindow returns the ±window (ms) for targets[i]: half the gap to the
// nearest neighbouring target at a different time, clamped to
// [minMatchWindowMs, maxMatchWindowMs]. Notes at the exact same time (e.g. the
// same drum on two note numbers) are skipped so they don't shrink the window.
func matchWindow(targets []models.NoteEvent, i int) float64 {
	t := targets[i].TimeMs
	gap := math.Inf(1)
	for j := i - 1; j >= 0; j-- {
		if d := t - targets[j].TimeMs; d > 0 {
			gap = d
			break
		}
	}
	for j := i + 1; j < len(targets); j++ {
		if d := targets[j].TimeMs - t; d > 0 {
			gap = math.Min(gap, d)
			break
		}
	}
	return math.Max(minMatchWindowMs, math.Min(maxMatchWindowMs, gap/2))
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
