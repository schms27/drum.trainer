import { g as getTrack } from './tracks-bMaygkn4.js';
import { e as error } from './index-CoD1IJuy.js';

const load = async ({ params }) => {
  const id = parseInt(params.trackId, 10);
  if (isNaN(id)) error(400, "Invalid track ID");
  try {
    const track = await getTrack(id);
    return { track };
  } catch (e) {
    const status = e?.status ?? 500;
    error(status, status === 404 ? "Track not found" : "Failed to load track");
  }
};

var _page_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 4;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BSDwt6uR.js')).default;
const universal_id = "src/routes/session/[trackId]/+page.ts";
const imports = ["_app/immutable/nodes/4.v8RP5QKa.js","_app/immutable/chunks/Cg2SOeN5.js","_app/immutable/chunks/uMvGNBtf.js","_app/immutable/chunks/CisNibZC.js","_app/immutable/chunks/BUApaBEI.js","_app/immutable/chunks/CfPFij4U.js","_app/immutable/chunks/CzMyd91B.js","_app/immutable/chunks/8sHoEGoU.js","_app/immutable/chunks/Bdi-rYIY.js","_app/immutable/chunks/BjUChR8d.js","_app/immutable/chunks/DBD2Xsyc.js","_app/immutable/chunks/D2e79A1n.js"];
const stylesheets = [];
const fonts = [];

export { component, fonts, imports, index, stylesheets, _page_ts as universal, universal_id };
//# sourceMappingURL=4-CfT-92XK.js.map
