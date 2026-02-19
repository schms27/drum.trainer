import { f as attr_class, h as stringify } from './index2-oMCSWlzO.js';

function MidiUpload($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    $$renderer2.push(`<div${attr_class(`border-2 border-dashed rounded-xl p-10 text-center transition-colors ${stringify("border-zinc-600 hover:border-zinc-400")}`)} role="region" aria-label="MIDI file drop zone">`);
    {
      $$renderer2.push("<!--[!-->");
      $$renderer2.push(`<p class="text-zinc-300 mb-1 text-lg">Drop a MIDI file here</p> <p class="text-zinc-500 text-sm mb-4">or click to browse</p> <label class="cursor-pointer inline-block px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm transition">Choose file <input type="file" accept=".mid,.midi" class="hidden"/></label>`);
    }
    $$renderer2.push(`<!--]--> `);
    {
      $$renderer2.push("<!--[!-->");
    }
    $$renderer2.push(`<!--]--></div>`);
  });
}
function TrackList($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    $$renderer2.push(`<div>`);
    {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<p class="text-zinc-500 text-sm">Loading tracks…</p>`);
    }
    $$renderer2.push(`<!--]--></div>`);
  });
}
function _page($$renderer) {
  $$renderer.push(`<div class="flex flex-col gap-10"><div><h1 class="text-2xl font-bold text-white mb-2">Upload a Target Track</h1> <p class="text-zinc-400 text-sm mb-5">Upload a standard MIDI file (.mid) to use as your practice target. The app will extract all
			drum notes from channel 10.</p> `);
  MidiUpload($$renderer);
  $$renderer.push(`<!----></div> <div><h2 class="text-xl font-semibold text-white mb-4">Your Tracks</h2> `);
  TrackList($$renderer);
  $$renderer.push(`<!----></div></div>`);
}

export { _page as default };
//# sourceMappingURL=_page.svelte-TJ8DEaRQ.js.map
