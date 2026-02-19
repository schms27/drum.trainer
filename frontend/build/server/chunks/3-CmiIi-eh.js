import { g as getTrack, a as apiFetch } from './tracks-bMaygkn4.js';
import { e as error } from './index-CoD1IJuy.js';

async function getSession(id) {
  return apiFetch(`/sessions/${id}`);
}
const load = async ({ params }) => {
  const id = parseInt(params.sessionId, 10);
  if (isNaN(id)) error(400, "Invalid session ID");
  try {
    const sessionResult = await getSession(id);
    const track = await getTrack(sessionResult.session.track_id);
    return { sessionResult, track };
  } catch (e) {
    const status = e?.status ?? 500;
    error(status, status === 404 ? "Session not found" : "Failed to load results");
  }
};

var _page_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 3;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-HX0wkW4I.js')).default;
const universal_id = "src/routes/results/[sessionId]/+page.ts";
const imports = ["_app/immutable/nodes/3.Bm_nsRwb.js","_app/immutable/chunks/CisNibZC.js","_app/immutable/chunks/Cg2SOeN5.js","_app/immutable/chunks/uMvGNBtf.js","_app/immutable/chunks/BUApaBEI.js","_app/immutable/chunks/CfPFij4U.js","_app/immutable/chunks/CzMyd91B.js","_app/immutable/chunks/TPvY_dZB.js","_app/immutable/chunks/8sHoEGoU.js","_app/immutable/chunks/Bdi-rYIY.js"];
const stylesheets = [];
const fonts = [];

export { component, fonts, imports, index, stylesheets, _page_ts as universal, universal_id };
//# sourceMappingURL=3-CmiIi-eh.js.map
