import { f as bind_props, b as ensure_array_like, a as attr_class, e as escape_html, h as clsx, d as derived, s as stringify, c as attr } from "../../../../chunks/index2.js";
import "@sveltejs/kit/internal";
import "../../../../chunks/exports.js";
import "../../../../chunks/utils.js";
import "clsx";
import "@sveltejs/kit/internal/server";
import "../../../../chunks/root.js";
import "../../../../chunks/state.svelte.js";
function MidiDeviceSelector($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { selectedId = "", onError } = $$props;
    $$renderer2.push(`<div class="flex flex-col gap-2"><label class="text-sm text-zinc-400 font-medium" for="midi-device">MIDI Input Device</label> `);
    {
      $$renderer2.push("<!--[1-->");
      $$renderer2.push(`<p class="text-zinc-500 text-sm">Scanning for MIDI devices…</p>`);
    }
    $$renderer2.push(`<!--]--></div>`);
    bind_props($$props, { selectedId });
  });
}
function LiveEventFeed($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    const GM_DRUM_MAP = {
      35: "Bass Drum",
      36: "Bass Drum",
      37: "Side Stick",
      38: "Snare",
      40: "Snare",
      42: "Hi-Hat (closed)",
      44: "Hi-Hat (pedal)",
      46: "Hi-Hat (open)",
      49: "Crash 1",
      57: "Crash 2",
      51: "Ride",
      53: "Ride Bell",
      59: "Ride 2",
      41: "Tom (lo floor)",
      43: "Tom (hi floor)",
      45: "Tom (low)",
      47: "Tom (lo-mid)",
      48: "Tom (hi-mid)",
      50: "Tom (high)"
    };
    let { hits } = $$props;
    const displayed = derived(() => hits.slice(-20).reverse());
    function label(note) {
      return GM_DRUM_MAP[note] ?? `Note ${note}`;
    }
    function velColor(vel) {
      if (vel > 90) return "text-green-400";
      if (vel > 50) return "text-yellow-400";
      return "text-zinc-400";
    }
    $$renderer2.push(`<div class="bg-zinc-800 rounded-xl p-4 border border-zinc-700 min-h-[160px]"><h3 class="text-xs text-zinc-500 uppercase tracking-wider mb-3">Live Feed</h3> `);
    if (displayed().length === 0) {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<p class="text-zinc-600 text-sm">No hits yet…</p>`);
    } else {
      $$renderer2.push("<!--[!-->");
      $$renderer2.push(`<ul class="space-y-1 text-sm font-mono"><!--[-->`);
      const each_array = ensure_array_like(displayed());
      for (let i = 0, $$length = each_array.length; i < $$length; i++) {
        let hit = each_array[i];
        $$renderer2.push(`<li${attr_class(`flex gap-3 items-center opacity-${stringify(i === 0 ? 100 : Math.max(40, 100 - i * 8))}`)}><span class="text-zinc-500 w-16 text-right">${escape_html(hit.time_ms.toFixed(0))}ms</span> <span class="text-white">${escape_html(label(hit.note_number))}</span> <span${attr_class(clsx(velColor(hit.velocity)))}>vel ${escape_html(hit.velocity)}</span></li>`);
      }
      $$renderer2.push(`<!--]--></ul>`);
    }
    $$renderer2.push(`<!--]--></div>`);
  });
}
function RecordingControls($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { deviceId } = $$props;
    let liveHits = [];
    $$renderer2.push(`<div class="flex flex-col gap-4"><div class="flex gap-3 items-center">`);
    {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<button${attr("disabled", !deviceId, true)} class="px-5 py-2 bg-green-600 hover:bg-green-500 disabled:bg-zinc-700 disabled:text-zinc-500 text-white rounded-lg font-medium transition">▶ Start Recording</button>`);
    }
    $$renderer2.push(`<!--]--> `);
    {
      $$renderer2.push("<!--[!-->");
    }
    $$renderer2.push(`<!--]--></div> `);
    {
      $$renderer2.push("<!--[!-->");
    }
    $$renderer2.push(`<!--]--> `);
    LiveEventFeed($$renderer2, { hits: liveHits });
    $$renderer2.push(`<!----></div>`);
  });
}
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    const { track } = data;
    let selectedDeviceId = "";
    function formatDuration(ms) {
      const secs = Math.floor(ms / 1e3);
      const m = Math.floor(secs / 60);
      const s = secs % 60;
      return `${m}:${String(s).padStart(2, "0")}`;
    }
    const noteCounts = derived(() => track.notes.reduce(
      (acc, n) => {
        acc[n.instrument] = (acc[n.instrument] ?? 0) + 1;
        return acc;
      },
      {}
    ));
    let $$settled = true;
    let $$inner_renderer;
    function $$render_inner($$renderer3) {
      $$renderer3.push(`<div class="flex flex-col gap-8"><div><a href="/" class="text-zinc-500 hover:text-zinc-300 text-sm transition">← Back to tracks</a> <h1 class="text-2xl font-bold text-white mt-2">${escape_html(track.name)}</h1> <p class="text-zinc-400 text-sm mt-1">${escape_html(track.note_count)} notes · ${escape_html(formatDuration(track.duration_ms))} · ${escape_html(track.filename)}</p></div> <div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 max-w-sm">`);
      MidiDeviceSelector($$renderer3, {
        get selectedId() {
          return selectedDeviceId;
        },
        set selectedId($$value) {
          selectedDeviceId = $$value;
          $$settled = false;
        }
      });
      $$renderer3.push(`<!----></div> <div><h2 class="text-lg font-semibold text-white mb-4">Recording</h2> `);
      {
        $$renderer3.push("<!--[!-->");
        RecordingControls($$renderer3, { deviceId: selectedDeviceId });
      }
      $$renderer3.push(`<!--]--> `);
      {
        $$renderer3.push("<!--[!-->");
      }
      $$renderer3.push(`<!--]--></div> <div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5"><h3 class="text-sm font-semibold text-zinc-300 mb-3">Target Track Contents</h3> <div class="flex flex-wrap gap-3 text-sm"><!--[-->`);
      const each_array = ensure_array_like(Object.entries(noteCounts()).sort((a, b) => b[1] - a[1]));
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let [instr, count] = each_array[$$index];
        $$renderer3.push(`<span class="bg-zinc-700 text-zinc-300 rounded-full px-3 py-1">${escape_html(instr)}: ${escape_html(count)}</span>`);
      }
      $$renderer3.push(`<!--]--></div></div></div>`);
    }
    do {
      $$settled = true;
      $$inner_renderer = $$renderer2.copy();
      $$render_inner($$inner_renderer);
    } while (!$$settled);
    $$renderer2.subsume($$inner_renderer);
  });
}
export {
  _page as default
};
