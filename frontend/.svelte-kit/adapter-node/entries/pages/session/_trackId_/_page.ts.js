import { g as getTrack } from "../../../../chunks/tracks.js";
import { error } from "@sveltejs/kit";
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
export {
  load
};
