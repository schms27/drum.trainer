# UI Overhaul Plan

Branch: `ui-overhaul`. Goal: make the app self-explanatory and easier to use.
Decisions below were settled on 2026-10-16.

**Status:** phases 1 and 2 implemented; phases 3–4 not started.

## Problems found (review of every page, desktop + phone)

1. **The practice flow isn't explained.** Nothing says recording starts on the first hit; "Start Recording" is disabled with the reason in a separate box; no next step after stopping.
2. **Library buried under upload.** Upload form fills the top half of the home page; the MIDI channel dropdown is jargon shown first.
3. **Results show numbers, not meaning.** No plain-language summary, no comparison with previous/best attempt, no suggested next step.
4. **Bugs and inconsistencies.**
   - Timeline captures the mouse wheel, so the page can't scroll while the pointer is over it.
   - Missed notes look identical to ordinary target circles.
   - Timeline always shows all nine lanes, even for a three-instrument track.
   - Metronome shows four beat dots regardless of time signature.
   - Histogram wastes most of its width on the empty ±200 ms range and clips its "+200ms" label.
   - Mapping and history pages use their own layouts.
   - Navigation doesn't highlight the current page.
5. **Phone/tablet widths break.** The header overflows (sign-out pushed off-screen) and the history table runs off the side.
6. **Jargon; progress is disconnected.** "Pilot tracks", "Ghost notes", "std deviation" go unexplained. Progress is spread over three sparse yearly charts, none showing accuracy over time.

## Decisions

| Topic | Decision |
|---|---|
| Recording start | **Count-in by default** (one bar of metronome; defines t=0). First-hit start stays available as a setting. |
| Pad mapping | **Kit profile + per-track overrides.** Map the kit once, with a "hit a pad to assign" learn mode. |
| History & Stats | **Merge into one "Progress" page.** Navigation becomes Library · Progress. |
| Track management | **Dedicated track page** (practice, best/recent sessions, backing tracks, mapping, delete). Library cards stay simple. |
| Theme | **Light + dark**, following the system setting with a toggle. |
| Icons | **lucide-svelte** replaces the emoji icons. |
| Wording | **Plain terms**: Backing tracks (Pilot tracks), Extra hits (Ghost notes), Timing spread (Consistency), Pad mapping (Mapping). |

### Technical consequences

- **Count-in makes latency matter again.** With first-hit start, a constant kit/USB delay cancels out; with a metronome-defined t=0 it shows up as "late" on every note. Add a short tap-along latency calibration, stored per device. MIDI `event.timeStamp` (performance clock) must be related to the AudioContext clock (`getOutputTimestamp`) so hits line up with the count-in.
- **Kit profile splits the mapping model.** Today one per-track mapping is applied to both the MIDI file's notes and the recorded hits. New model: the **kit profile maps recorded hits** (your pads), the **per-track mapping maps the file's notes**. Needs a `kit_profiles` table, a user's active profile, and scoring changes (hit → kit profile → GM default; target → track mapping → GM default).
- **Light + dark needs design tokens first.** Every component hard-codes `zinc-*` classes. Introduce semantic color tokens (Tailwind 4 `@theme` + CSS variables, dark/light values) before restyling pages.

## Phases

### Phase 1 — Foundation and quick fixes (frontend only) ✅
- Semantic color tokens with light/dark themes and a theme toggle.
- Shared components: page container, button, card, dialog (replaces `confirm()`), empty state; lucide icons.
- Header: active-page highlight, mobile menu, Library · Progress navigation (Progress initially links the existing pages).
- Plain-language terms everywhere.
- Timeline: hide unused lanes, stop hijacking page scroll, distinct missed-note marker.
- Histogram zooms to the data range; fix the clipped label.
- Metronome beat dots follow the time signature.
- Practice page states why Start is disabled, inline.
- History table becomes cards on narrow screens.

### Phase 2 — Guided practice flow ✅
- Set up → Play → Review layout on the practice page.
- Device check with live "hit a pad" indicator; device list refreshes when a kit is plugged in.
- Speed presets (50/75/90/100%) next to the BPM control.
- Count-in (default) and first-hit start (setting); latency calibration.
- Large status banner, automatic stop at track end, focus mode.

### Phase 3 — Progress and track pages (needs API work)
- Track list/detail API with best score, session count, last practiced, hit-rate trend.
- Dedicated track page; simplified library cards with automatic drum channel detection (channel picker under "Advanced").
- Results: one-sentence summary, change versus last attempt and best, "Try again" / "Try slower" actions, metric explanations.
- Progress page merging History and Stats: accuracy trend per track, practice time and streak, session list with track filter.

### Phase 4 — Kit profiles and onboarding
- Kit profiles with learn mode; per-track overrides; scoring changes as described above.
- First-run onboarding: requirements (Chrome/Edge, USB kit, HTTPS) and the three-step flow.
