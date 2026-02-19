import "clsx";
function Header($$renderer) {
  $$renderer.push(`<header class="bg-zinc-900 border-b border-zinc-700 px-6 py-4 flex items-center gap-4"><a href="/" class="text-white font-bold text-xl tracking-tight hover:text-indigo-400 transition">🥁 Drum Trainer</a> <nav class="flex gap-4 ml-auto text-sm"><a href="/" class="text-zinc-400 hover:text-white transition">Tracks</a></nav></header>`);
}
function _layout($$renderer, $$props) {
  let { children } = $$props;
  $$renderer.push(`<div class="min-h-screen bg-zinc-950 text-white">`);
  Header($$renderer);
  $$renderer.push(`<!----> <main class="max-w-5xl mx-auto px-4 py-8">`);
  children($$renderer);
  $$renderer.push(`<!----></main></div>`);
}
export {
  _layout as default
};
