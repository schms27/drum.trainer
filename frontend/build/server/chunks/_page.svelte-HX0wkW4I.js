import { j as attr, h as stringify, c as escape_html, f as attr_class, k as ensure_array_like, l as derived } from './index2-oMCSWlzO.js';
import * as d3 from 'd3';

function ScoreSummary($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { score } = $$props;
    function pct(v) {
      return (v * 100).toFixed(1) + "%";
    }
    function offsetLabel(ms) {
      if (Math.abs(ms) < 0.5) return "±0 ms";
      return (ms > 0 ? "+" : "") + ms.toFixed(1) + " ms";
    }
    function offsetColor(ms) {
      const abs = Math.abs(ms);
      if (abs <= 20) return "text-green-400";
      if (abs <= 50) return "text-yellow-400";
      return "text-red-400";
    }
    function hitRateColor(r) {
      if (r >= 0.9) return "text-green-400";
      if (r >= 0.7) return "text-yellow-400";
      return "text-red-400";
    }
    $$renderer2.push(`<div class="grid grid-cols-2 sm:grid-cols-4 gap-4"><div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 text-center"><p class="text-zinc-400 text-xs uppercase tracking-wider mb-1">Hit Rate</p> <p${attr_class(`text-3xl font-bold ${stringify(hitRateColor(score.hit_rate))}`)}>${escape_html(pct(score.hit_rate))}</p></div> <div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 text-center"><p class="text-zinc-400 text-xs uppercase tracking-wider mb-1">Avg Timing</p> <p${attr_class(`text-3xl font-bold ${stringify(offsetColor(score.avg_offset_ms))}`)}>${escape_html(offsetLabel(score.avg_offset_ms))}</p> <p class="text-zinc-500 text-xs mt-1">${escape_html(score.avg_offset_ms < 0 ? "playing early" : score.avg_offset_ms > 0 ? "playing late" : "perfect")}</p></div> <div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 text-center"><p class="text-zinc-400 text-xs uppercase tracking-wider mb-1">Consistency</p> <p class="text-3xl font-bold text-white">±${escape_html(score.std_offset_ms.toFixed(1))} ms</p> <p class="text-zinc-500 text-xs mt-1">std deviation</p></div> <div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 text-center"><p class="text-zinc-400 text-xs uppercase tracking-wider mb-1">Ghost Notes</p> <p class="text-3xl font-bold text-zinc-300">${escape_html(pct(score.ghost_rate))}</p></div></div>`);
  });
}
function InstrumentTable($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { breakdown } = $$props;
    function pct(v) {
      return (v * 100).toFixed(1) + "%";
    }
    function hitColor(r) {
      if (r >= 0.9) return "text-green-400";
      if (r >= 0.7) return "text-yellow-400";
      return "text-red-400";
    }
    function offsetStr(v) {
      if (v === null) return "—";
      return (v > 0 ? "+" : "") + v.toFixed(1) + " ms";
    }
    function offsetColor(v) {
      if (v === null) return "text-zinc-500";
      const abs = Math.abs(v);
      if (abs <= 20) return "text-green-400";
      if (abs <= 50) return "text-yellow-400";
      return "text-red-400";
    }
    $$renderer2.push(`<div class="overflow-x-auto"><table class="w-full text-sm"><thead><tr class="text-zinc-500 text-xs uppercase tracking-wider border-b border-zinc-700"><th class="text-left py-2 pr-4">Instrument</th><th class="text-right py-2 px-3">Target</th><th class="text-right py-2 px-3">Hit</th><th class="text-right py-2 px-3">Miss</th><th class="text-right py-2 px-3">Ghost</th><th class="text-right py-2 px-3">Hit Rate</th><th class="text-right py-2 pl-3">Avg Offset</th></tr></thead><tbody><!--[-->`);
    const each_array = ensure_array_like(breakdown);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let row = each_array[$$index];
      $$renderer2.push(`<tr class="border-b border-zinc-800 hover:bg-zinc-800/50 transition"><td class="py-2.5 pr-4 text-white font-medium">${escape_html(row.instrument)}</td><td class="text-right py-2.5 px-3 text-zinc-400">${escape_html(row.target_count)}</td><td class="text-right py-2.5 px-3 text-zinc-300">${escape_html(row.matched_count)}</td><td class="text-right py-2.5 px-3 text-red-400">${escape_html(row.missed_count)}</td><td class="text-right py-2.5 px-3 text-zinc-400">${escape_html(row.ghost_count)}</td><td${attr_class(`text-right py-2.5 px-3 font-medium ${stringify(hitColor(row.hit_rate))}`)}>${escape_html(pct(row.hit_rate))}</td><td${attr_class(`text-right py-2.5 pl-3 font-mono ${stringify(offsetColor(row.avg_offset_ms))}`)}>${escape_html(offsetStr(row.avg_offset_ms))}</td></tr>`);
    }
    $$renderer2.push(`<!--]--></tbody></table></div>`);
  });
}
function TimingHistogram($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { hits } = $$props;
    const WIDTH = 560;
    const HEIGHT = 200;
    const MARGIN = { top: 16, right: 16, bottom: 36, left: 40 };
    const DOMAIN = [-200, 200];
    const BIN_COUNT = 20;
    const innerW = WIDTH - MARGIN.left - MARGIN.right;
    const innerH = HEIGHT - MARGIN.top - MARGIN.bottom;
    const offsets = derived(() => hits.map((h) => h.offset_ms));
    const x = derived(() => d3.scaleLinear().domain(DOMAIN).range([0, innerW]));
    const bins = derived(() => d3.bin().domain(DOMAIN).thresholds(BIN_COUNT)(offsets()));
    const yMax = derived(() => d3.max(bins(), (b) => b.length) ?? 1);
    const y = derived(() => d3.scaleLinear().domain([0, yMax()]).range([innerH, 0]).nice());
    const xTicks = derived(() => x().ticks(5));
    const yTicks = derived(() => y().ticks(4));
    function barColor(x0, x1) {
      const center = ((x0 ?? 0) + (x1 ?? 0)) / 2;
      const abs = Math.abs(center);
      if (abs <= 20) return "#4ade80";
      if (abs <= 50) return "#facc15";
      return "#f87171";
    }
    $$renderer2.push(`<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5"><h3 class="text-sm font-semibold text-zinc-300 mb-4">Timing Distribution</h3> `);
    if (hits.length === 0) {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<p class="text-zinc-500 text-sm">No matched hits to display.</p>`);
    } else {
      $$renderer2.push("<!--[!-->");
      $$renderer2.push(`<svg${attr("viewBox", `0 0 ${stringify(WIDTH)} ${stringify(HEIGHT)}`)} class="w-full" aria-label="Timing histogram"><g${attr("transform", `translate(${stringify(MARGIN.left)},${stringify(MARGIN.top)})`)}><!--[-->`);
      const each_array = ensure_array_like(bins());
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let bin = each_array[$$index];
        if (bin.x0 !== void 0 && bin.x1 !== void 0) {
          $$renderer2.push("<!--[-->");
          $$renderer2.push(`<rect${attr("x", x()(bin.x0) + 1)}${attr("y", y()(bin.length))}${attr("width", Math.max(0, x()(bin.x1) - x()(bin.x0) - 2))}${attr("height", innerH - y()(bin.length))}${attr("fill", barColor(bin.x0, bin.x1))} opacity="0.85"></rect>`);
        } else {
          $$renderer2.push("<!--[!-->");
        }
        $$renderer2.push(`<!--]-->`);
      }
      $$renderer2.push(`<!--]--><line${attr("x1", x()(0))}${attr("y1", 0)}${attr("x2", x()(0))}${attr("y2", innerH)} stroke="#71717a" stroke-width="1" stroke-dasharray="4 3"></line><line${attr("x1", 0)}${attr("y1", innerH)}${attr("x2", innerW)}${attr("y2", innerH)} stroke="#52525b"></line><!--[-->`);
      const each_array_1 = ensure_array_like(xTicks());
      for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
        let tick = each_array_1[$$index_1];
        $$renderer2.push(`<text${attr("x", x()(tick))}${attr("y", innerH + 18)} text-anchor="middle" fill="#a1a1aa" font-size="11">${escape_html(tick > 0 ? `+${tick}` : tick)}ms</text>`);
      }
      $$renderer2.push(`<!--]--><!--[-->`);
      const each_array_2 = ensure_array_like(yTicks());
      for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
        let tick = each_array_2[$$index_2];
        $$renderer2.push(`<line${attr("x1", 0)}${attr("y1", y()(tick))}${attr("x2", innerW)}${attr("y2", y()(tick))} stroke="#3f3f46"></line><text${attr("x", -6)}${attr("y", y()(tick) + 4)} text-anchor="end" fill="#71717a" font-size="10">${escape_html(tick)}</text>`);
      }
      $$renderer2.push(`<!--]--><text${attr("x", innerW / 2)}${attr("y", innerH + 34)} text-anchor="middle" fill="#71717a" font-size="11">Timing offset (ms)</text></g></svg> <div class="flex gap-4 mt-2 justify-center text-xs text-zinc-500"><span><span class="inline-block w-3 h-3 rounded-sm bg-green-400 mr-1"></span>±20 ms</span> <span><span class="inline-block w-3 h-3 rounded-sm bg-yellow-400 mr-1"></span>±50 ms</span> <span><span class="inline-block w-3 h-3 rounded-sm bg-red-400 mr-1"></span>>50 ms</span></div>`);
    }
    $$renderer2.push(`<!--]--></div>`);
  });
}
function TimelineView($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { targetNotes, matchedHits, breakdown, durationMs } = $$props;
    const WIDTH = 680;
    const ROW_H = 32;
    const LABEL_W = 90;
    const MARGIN = { top: 8, bottom: 16 };
    const DOT_R = 5;
    const groups = derived(() => breakdown.map((b) => b.instrument));
    const svgH = derived(() => MARGIN.top + groups().length * ROW_H + MARGIN.bottom);
    const trackW = derived(() => WIDTH - LABEL_W);
    const msPerPx = derived(() => durationMs > 0 ? durationMs / trackW() : 1);
    function xOf(ms) {
      return LABEL_W + ms / msPerPx();
    }
    function rowY(i) {
      return MARGIN.top + i * ROW_H + ROW_H / 2;
    }
    function offsetDotColor(offset) {
      const a = Math.abs(offset);
      if (a <= 20) return "#4ade80";
      if (a <= 50) return "#facc15";
      return "#f87171";
    }
    $$renderer2.push(`<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 overflow-x-auto"><h3 class="text-sm font-semibold text-zinc-300 mb-4">Timeline</h3> `);
    if (targetNotes.length === 0) {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<p class="text-zinc-500 text-sm">No data.</p>`);
    } else {
      $$renderer2.push("<!--[!-->");
      $$renderer2.push(`<svg${attr("viewBox", `0 0 ${stringify(WIDTH)} ${stringify(svgH())}`)} class="w-full min-w-[400px]" aria-label="Hit timeline"><!--[-->`);
      const each_array = ensure_array_like(groups());
      for (let i = 0, $$length = each_array.length; i < $$length; i++) {
        let group = each_array[i];
        $$renderer2.push(`<rect${attr("x", 0)}${attr("y", MARGIN.top + i * ROW_H)}${attr("width", WIDTH)}${attr("height", ROW_H)}${attr("fill", i % 2 === 0 ? "#27272a" : "#1c1c1e")}></rect><text${attr("x", LABEL_W - 8)}${attr("y", rowY(i) + 4)} text-anchor="end" fill="#a1a1aa" font-size="11">${escape_html(group)}</text><line${attr("x1", LABEL_W)}${attr("y1", MARGIN.top + (i + 1) * ROW_H)}${attr("x2", WIDTH)}${attr("y2", MARGIN.top + (i + 1) * ROW_H)} stroke="#3f3f46"></line>`);
      }
      $$renderer2.push(`<!--]--><!--[-->`);
      const each_array_1 = ensure_array_like(targetNotes);
      for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
        let note = each_array_1[$$index_1];
        const gi = groups().indexOf(note.instrument);
        if (gi >= 0) {
          $$renderer2.push("<!--[-->");
          $$renderer2.push(`<circle${attr("cx", xOf(note.time_ms))}${attr("cy", rowY(gi))}${attr("r", DOT_R)} fill="none" stroke="#6366f1" stroke-width="1.5" opacity="0.7"></circle>`);
        } else {
          $$renderer2.push("<!--[!-->");
        }
        $$renderer2.push(`<!--]-->`);
      }
      $$renderer2.push(`<!--]--><!--[-->`);
      const each_array_2 = ensure_array_like(matchedHits);
      for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
        let hit = each_array_2[$$index_2];
        const gi = groups().indexOf(hit.instrument);
        if (gi >= 0) {
          $$renderer2.push("<!--[-->");
          $$renderer2.push(`<circle${attr("cx", xOf(hit.recorded_time_ms))}${attr("cy", rowY(gi))}${attr("r", DOT_R - 1)}${attr("fill", offsetDotColor(hit.offset_ms))} opacity="0.9"><title>${escape_html(hit.instrument)} | offset: ${escape_html(hit.offset_ms.toFixed(1))}ms</title></circle>`);
        } else {
          $$renderer2.push("<!--[!-->");
        }
        $$renderer2.push(`<!--]-->`);
      }
      $$renderer2.push(`<!--]--><line${attr("x1", LABEL_W)}${attr("y1", 0)}${attr("x2", LABEL_W)}${attr("y2", svgH())} stroke="#52525b"></line></svg> <div class="flex gap-5 mt-3 text-xs text-zinc-500 justify-center"><span><svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10"><circle cx="5" cy="5" r="4" fill="none" stroke="#6366f1" stroke-width="1.5"></circle></svg>Target</span> <span><svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10"><circle cx="5" cy="5" r="3" fill="#4ade80"></circle></svg>Hit (±20ms)</span> <span><svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10"><circle cx="5" cy="5" r="3" fill="#facc15"></circle></svg>Hit (±50ms)</span> <span><svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10"><circle cx="5" cy="5" r="3" fill="#f87171"></circle></svg>Late/Early</span></div>`);
    }
    $$renderer2.push(`<!--]--></div>`);
  });
}
function _page($$renderer, $$props) {
  let { data } = $$props;
  const { sessionResult, track } = data;
  const { session, score } = sessionResult;
  $$renderer.push(`<div class="flex flex-col gap-8"><div><a${attr("href", `/session/${stringify(session.track_id)}`)} class="text-zinc-500 hover:text-zinc-300 text-sm transition">← Practice again</a> <h1 class="text-2xl font-bold text-white mt-2">Results — ${escape_html(track.name)}</h1> <p class="text-zinc-400 text-sm mt-1">${escape_html(session.hit_count)} hits recorded · Session #${escape_html(session.id)}</p></div> `);
  ScoreSummary($$renderer, { score });
  $$renderer.push(`<!----> <div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5"><h2 class="text-sm font-semibold text-zinc-300 mb-4">Per-Instrument Breakdown</h2> `);
  InstrumentTable($$renderer, { breakdown: score.by_instrument });
  $$renderer.push(`<!----></div> `);
  TimingHistogram($$renderer, { hits: score.matched_hits });
  $$renderer.push(`<!----> `);
  TimelineView($$renderer, {
    targetNotes: track.notes,
    matchedHits: score.matched_hits,
    breakdown: score.by_instrument,
    durationMs: track.duration_ms
  });
  $$renderer.push(`<!----></div>`);
}

export { _page as default };
//# sourceMappingURL=_page.svelte-HX0wkW4I.js.map
