const BASE = "/api/v1";
class ApiError extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
    this.name = "ApiError";
  }
}
async function apiFetch(path, init) {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...init?.headers },
    ...init
  });
  if (!res.ok) {
    let message = `HTTP ${res.status}`;
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
    } catch {
    }
    throw new ApiError(res.status, message);
  }
  if (res.status === 204) return void 0;
  return res.json();
}
async function getTrack(id) {
  return apiFetch(`/tracks/${id}`);
}

export { apiFetch as a, getTrack as g };
//# sourceMappingURL=tracks-bMaygkn4.js.map
