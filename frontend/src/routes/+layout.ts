// Run all load functions in the browser (SPA mode).
// Relative API fetches like /api/v1/... only work client-side where the
// Vite dev proxy (and the production reverse proxy) can forward them to the backend.
export const ssr = false;
