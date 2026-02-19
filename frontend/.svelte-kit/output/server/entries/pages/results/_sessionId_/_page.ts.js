import { a as apiFetch, g as getTrack } from "../../../../chunks/tracks.js";
import { error } from "@sveltejs/kit";
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
export {
  load
};
