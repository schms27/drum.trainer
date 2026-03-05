package scoring_test

import (
	"math"
	"testing"

	"github.com/simu/drum-trainer/backend/models"
	"github.com/simu/drum-trainer/backend/scoring"
)

// --- helpers ---

func singleGroup(_ int) string { return "Drum" }

func kickOrSnare(n int) string {
	if n == 36 {
		return "Bass Drum"
	}
	return "Snare"
}

func note(n int, t float64) models.NoteEvent {
	return models.NoteEvent{NoteNumber: n, TimeMs: t}
}

func hit(n int, t float64) models.RecordedHit {
	return models.RecordedHit{NoteNumber: n, TimeMs: t}
}

func approxEqual(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}

// --- empty / nil input ---

func TestScore_BothEmpty(t *testing.T) {
	r := scoring.Score(nil, nil, singleGroup)
	if r.HitRate != 0 || r.MissRate != 0 || r.GhostRate != 0 {
		t.Errorf("expected all zeros, got hit=%.2f miss=%.2f ghost=%.2f", r.HitRate, r.MissRate, r.GhostRate)
	}
	if len(r.MatchedHits) != 0 || len(r.GhostHits) != 0 {
		t.Error("expected empty matched and ghost slices")
	}
}

func TestScore_NoHits_AllMissed(t *testing.T) {
	targets := []models.NoteEvent{note(36, 1000), note(36, 2000)}
	r := scoring.Score(targets, nil, singleGroup)
	if r.HitRate != 0 {
		t.Errorf("hit rate should be 0, got %.2f", r.HitRate)
	}
	if r.MissRate != 1.0 {
		t.Errorf("miss rate should be 1.0, got %.2f", r.MissRate)
	}
	if r.GhostRate != 0 {
		t.Errorf("ghost rate should be 0, got %.2f", r.GhostRate)
	}
}

func TestScore_NoTargets_AllGhosts(t *testing.T) {
	hits := []models.RecordedHit{hit(36, 1000), hit(36, 2000)}
	r := scoring.Score(nil, hits, singleGroup)
	if r.GhostRate != 1.0 {
		t.Errorf("ghost rate should be 1.0, got %.2f", r.GhostRate)
	}
	if r.HitRate != 0 || r.MissRate != 0 {
		t.Errorf("hit/miss should be 0, got hit=%.2f miss=%.2f", r.HitRate, r.MissRate)
	}
	if len(r.GhostHits) != 2 {
		t.Errorf("expected 2 ghost hits, got %d", len(r.GhostHits))
	}
}

// --- matching ---

func TestScore_PerfectMatch(t *testing.T) {
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)},
		[]models.RecordedHit{hit(36, 1000)},
		singleGroup,
	)
	if r.HitRate != 1.0 {
		t.Errorf("expected hit rate 1.0, got %.2f", r.HitRate)
	}
	if len(r.MatchedHits) != 1 {
		t.Fatalf("expected 1 matched hit, got %d", len(r.MatchedHits))
	}
	if r.MatchedHits[0].OffsetMs != 0 {
		t.Errorf("expected offset 0, got %.2f", r.MatchedHits[0].OffsetMs)
	}
	if r.AvgOffsetMs != 0 {
		t.Errorf("expected avg offset 0, got %.2f", r.AvgOffsetMs)
	}
}

func TestScore_HitLateWithinWindow(t *testing.T) {
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)},
		[]models.RecordedHit{hit(36, 1100)},
		singleGroup,
	)
	if r.HitRate != 1.0 {
		t.Errorf("hit +100ms should match, got hit rate %.2f", r.HitRate)
	}
	if r.MatchedHits[0].OffsetMs != 100 {
		t.Errorf("expected offset +100, got %.2f", r.MatchedHits[0].OffsetMs)
	}
}

func TestScore_HitEarlyWithinWindow(t *testing.T) {
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)},
		[]models.RecordedHit{hit(36, 900)},
		singleGroup,
	)
	if r.HitRate != 1.0 {
		t.Errorf("hit -100ms should match, got hit rate %.2f", r.HitRate)
	}
	if r.MatchedHits[0].OffsetMs != -100 {
		t.Errorf("expected offset -100, got %.2f", r.MatchedHits[0].OffsetMs)
	}
}

func TestScore_HitAtLateWindowBoundary(t *testing.T) {
	// Exactly +200ms should still match (delta <= matchWindowMs).
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)},
		[]models.RecordedHit{hit(36, 1200)},
		singleGroup,
	)
	if r.HitRate != 1.0 {
		t.Errorf("hit at +200ms boundary should match, got hit rate %.2f", r.HitRate)
	}
}

func TestScore_HitAtEarlyWindowBoundary(t *testing.T) {
	// Exactly -200ms should still match.
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)},
		[]models.RecordedHit{hit(36, 800)},
		singleGroup,
	)
	if r.HitRate != 1.0 {
		t.Errorf("hit at -200ms boundary should match, got hit rate %.2f", r.HitRate)
	}
}

func TestScore_HitJustOutsideWindowLate(t *testing.T) {
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)},
		[]models.RecordedHit{hit(36, 1201)},
		singleGroup,
	)
	if r.HitRate != 0 {
		t.Errorf("hit at +201ms should miss, got hit rate %.2f", r.HitRate)
	}
	if r.MissRate != 1.0 {
		t.Errorf("expected miss rate 1.0, got %.2f", r.MissRate)
	}
	if r.GhostRate != 1.0 {
		t.Errorf("expected ghost rate 1.0, got %.2f", r.GhostRate)
	}
}

func TestScore_HitJustOutsideWindowEarly(t *testing.T) {
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)},
		[]models.RecordedHit{hit(36, 799)},
		singleGroup,
	)
	if r.HitRate != 0 {
		t.Errorf("hit at -201ms should miss, got hit rate %.2f", r.HitRate)
	}
}

// --- greedy matching ---

func TestScore_NearestHitChosen(t *testing.T) {
	// Two hits near one target; the nearer one should be used.
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)},
		[]models.RecordedHit{hit(36, 950), hit(36, 1080)},
		singleGroup,
	)
	if r.HitRate != 1.0 {
		t.Fatalf("expected 1 match, got hit rate %.2f", r.HitRate)
	}
	// hit at 950ms is closer (|delta|=50) than 1080ms (|delta|=80)
	if r.MatchedHits[0].OffsetMs != -50 {
		t.Errorf("expected nearer hit matched (offset -50), got %.2f", r.MatchedHits[0].OffsetMs)
	}
	if len(r.GhostHits) != 1 {
		t.Errorf("expected 1 ghost, got %d", len(r.GhostHits))
	}
}

func TestScore_OneHitMultipleTargets(t *testing.T) {
	// One hit for two targets — only the closer target is matched.
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000), note(36, 2000)},
		[]models.RecordedHit{hit(36, 1000)},
		singleGroup,
	)
	if r.HitRate != 0.5 {
		t.Errorf("expected 0.5 hit rate, got %.2f", r.HitRate)
	}
	if r.MissRate != 0.5 {
		t.Errorf("expected 0.5 miss rate, got %.2f", r.MissRate)
	}
	if len(r.GhostHits) != 0 {
		t.Errorf("expected no ghosts, got %d", len(r.GhostHits))
	}
}

func TestScore_HitNotReuseAcrossTargets(t *testing.T) {
	// Two targets at similar times, one hit — each hit can only be used once.
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000), note(36, 1050)},
		[]models.RecordedHit{hit(36, 1000)},
		singleGroup,
	)
	if len(r.MatchedHits) != 1 {
		t.Errorf("one hit should match at most one target, got %d matches", len(r.MatchedHits))
	}
	if len(r.GhostHits) != 0 {
		t.Errorf("used hit should not be ghost, got %d ghosts", len(r.GhostHits))
	}
}

// --- instrument group isolation ---

func TestScore_InstrumentGroupIsolation(t *testing.T) {
	// Kick target at the same time as a snare hit — should not match.
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)}, // kick target
		[]models.RecordedHit{hit(38, 1000)}, // snare hit
		kickOrSnare,
	)
	if r.HitRate != 0 {
		t.Errorf("kick should not be matched by snare hit, got hit rate %.2f", r.HitRate)
	}
	if len(r.GhostHits) != 1 {
		t.Errorf("snare hit should be ghost, got %d ghosts", len(r.GhostHits))
	}
}

func TestScore_MultipleGroups_IndependentMatching(t *testing.T) {
	targets := []models.NoteEvent{
		note(36, 1000), // kick
		note(38, 1000), // snare
	}
	hits := []models.RecordedHit{
		hit(36, 990), // kick hit
		hit(38, 1010), // snare hit
	}
	r := scoring.Score(targets, hits, kickOrSnare)
	if r.HitRate != 1.0 {
		t.Errorf("both instruments should match, got hit rate %.2f", r.HitRate)
	}
	if len(r.MatchedHits) != 2 {
		t.Errorf("expected 2 matches, got %d", len(r.MatchedHits))
	}
}

// --- statistics ---

func TestScore_OffsetAverage_Symmetric(t *testing.T) {
	// +100ms and -100ms → mean=0, std>0
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000), note(36, 2000)},
		[]models.RecordedHit{hit(36, 1100), hit(36, 1900)},
		singleGroup,
	)
	if !approxEqual(r.AvgOffsetMs, 0, 0.001) {
		t.Errorf("expected avg offset 0, got %.4f", r.AvgOffsetMs)
	}
	if r.StdOffsetMs <= 0 {
		t.Error("expected positive std dev for asymmetric hits")
	}
}

func TestScore_OffsetAverage_AllLate(t *testing.T) {
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000), note(36, 2000)},
		[]models.RecordedHit{hit(36, 1050), hit(36, 2050)},
		singleGroup,
	)
	if !approxEqual(r.AvgOffsetMs, 50, 0.001) {
		t.Errorf("expected avg offset 50ms, got %.4f", r.AvgOffsetMs)
	}
}

func TestScore_StdDev_Zeros(t *testing.T) {
	// Single match → stddev = 0 (by convention for n<2)
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)},
		[]models.RecordedHit{hit(36, 1050)},
		singleGroup,
	)
	if r.StdOffsetMs != 0 {
		t.Errorf("expected std dev 0 for single match, got %.4f", r.StdOffsetMs)
	}
}

func TestScore_GhostRate(t *testing.T) {
	// 1 target, 3 hits: 1 matched, 2 ghosts → ghost rate = 2/3
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)},
		[]models.RecordedHit{hit(36, 1000), hit(36, 3000), hit(36, 4000)},
		singleGroup,
	)
	want := 2.0 / 3.0
	if !approxEqual(r.GhostRate, want, 0.001) {
		t.Errorf("expected ghost rate %.3f, got %.3f", want, r.GhostRate)
	}
}

// --- per-instrument breakdown ---

func TestScore_ByInstrument_HitRate(t *testing.T) {
	targets := []models.NoteEvent{
		note(36, 1000), // kick — will miss
		note(38, 1000), // snare — will hit
	}
	hits := []models.RecordedHit{
		hit(38, 1000), // only snare matched
	}
	r := scoring.Score(targets, hits, kickOrSnare)

	bd := make(map[string]float64)
	for _, b := range r.ByInstrument {
		bd[b.Instrument] = b.HitRate
	}
	if bd["Bass Drum"] != 0 {
		t.Errorf("kick hit rate should be 0, got %.2f", bd["Bass Drum"])
	}
	if bd["Snare"] != 1.0 {
		t.Errorf("snare hit rate should be 1.0, got %.2f", bd["Snare"])
	}
}

func TestScore_ByInstrument_OffsetNilWhenNoMatches(t *testing.T) {
	// Kick missed → its AvgOffsetMs/StdOffsetMs should be nil.
	r := scoring.Score(
		[]models.NoteEvent{note(36, 1000)},
		nil,
		kickOrSnare,
	)
	for _, b := range r.ByInstrument {
		if b.Instrument == "Bass Drum" {
			if b.AvgOffsetMs != nil || b.StdOffsetMs != nil {
				t.Error("expected nil offset stats for instrument with no matches")
			}
		}
	}
}
