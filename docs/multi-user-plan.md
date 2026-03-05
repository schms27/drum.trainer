# Multi-User Support — Implementation Plan

## Overview

Add user accounts and data isolation to the drum trainer. Users authenticate via OAuth providers (Google, GitHub) — no password database. Data (tracks, sessions) is scoped per user.

**Auth mechanism:** Server-side sessions stored in the DB, referenced via a signed HttpOnly cookie. Immune to XSS; no localStorage JWTs.

---

## Step 1 — New config env vars

`backend/config/config.go` gains the following fields:

| Variable | Purpose |
|---|---|
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | Google OAuth2 app credentials |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | GitHub OAuth2 app credentials |
| `SESSION_SECRET` | 32-byte random hex, used to HMAC-sign state cookies |
| `BASE_URL` | Public base URL (e.g. `https://drums.example.com`) — needed for OAuth callback URLs |

---

## Step 2 — DB migrations

### `007_add_users.sql` (sqlite + postgres variants)

```sql
CREATE TABLE users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,  -- SERIAL in postgres
    provider    TEXT    NOT NULL,   -- "google" | "github"
    provider_id TEXT    NOT NULL,   -- subject claim from OAuth token
    email       TEXT    NOT NULL,
    name        TEXT    NOT NULL,
    avatar_url  TEXT,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX idx_users_provider ON users(provider, provider_id);

CREATE TABLE user_sessions (
    token       TEXT    PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TEXT    NOT NULL,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);
```

### `008_add_user_id.sql`

Add ownership columns to existing tables. Nullable first to allow safe migration of any existing rows.

```sql
ALTER TABLE tracks   ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE sessions ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
```

Existing dev rows will have `user_id = NULL` and will not appear for any authenticated user — acceptable for dev. On a fresh production DB all rows are created with a user_id from day one.

---

## Step 3 — New backend packages

### `backend/auth/`

| File | Responsibility |
|---|---|
| `oauth.go` | Build `*oauth2.Config` for each provider; define scopes; fetch userinfo from provider API |
| `session.go` | `CreateSession(db, userID)` — generates 32-byte random token, stores in `user_sessions` with 30-day expiry; `ValidateSession(db, token)` — returns `*models.User`; `DeleteSession(db, token)` |
| `state.go` | CSRF state helpers: generate + set a short-lived signed state cookie, verify on callback |

New Go dependency: `golang.org/x/oauth2` (includes `golang.org/x/oauth2/google` and `golang.org/x/oauth2/github` sub-packages).

### `backend/db/users.go`

```go
UpsertUser(db, provider, providerID, email, name, avatarURL) (*models.User, error)
GetUserByID(db, id) (*models.User, error)
```

### `backend/models/user.go`

```go
type User struct {
    ID        int64
    Provider  string
    Email     string
    Name      string
    AvatarURL string
    CreatedAt time.Time
}
```

---

## Step 4 — New API handlers and routes

### `backend/api/auth.go`

| Handler | Route | What it does |
|---|---|---|
| `loginHandler` | `GET /api/v1/auth/{provider}/login` | Generates CSRF state cookie, redirects to provider OAuth URL |
| `callbackHandler` | `GET /api/v1/auth/{provider}/callback` | Validates state, exchanges code for token, fetches userinfo, upserts user, sets session cookie, redirects to `/` |
| `logoutHandler` | `POST /api/v1/auth/logout` | Deletes session from DB, clears cookie, returns 204 |
| `meHandler` | `GET /api/v1/auth/me` | Returns current user as JSON (used by frontend on every page load) |

Session cookie spec:
```
Name:     drum_session
Flags:    HttpOnly; Secure; SameSite=Lax; Path=/
Max-Age:  2592000  (30 days)
```

In development (HTTP localhost) the `Secure` flag can be relaxed via a `DEV_MODE=true` env var.

### `backend/api/router.go` changes

- Mount auth routes outside the auth-required group (unauthenticated access needed for login/callback)
- Wrap `/api/v1/tracks` and `/api/v1/sessions` subrouters with `RequireAuth` middleware
- Pass auth config into `NewRouter`
- CORS: set `AllowCredentials: true` and allow the cookie origin

---

## Step 5 — `RequireAuth` middleware

`backend/api/middleware.go` gains:

```go
func RequireAuth(db *sql.DB) func(http.Handler) http.Handler
```

Logic:
1. Read `drum_session` cookie from request
2. Call `auth.ValidateSession(db, token)`
3. On success: inject `*models.User` into `context` via a typed key
4. On failure (missing, expired, unknown token): return `401 {"error": "unauthorized"}`

Helper exposed to handlers: `UserFromContext(ctx context.Context) *models.User`

---

## Step 6 — Scope all DB queries by user

### `backend/db/tracks.go`

Every function gains a `userID int64` parameter:

| Function | Change |
|---|---|
| `ListTracks` | Add `WHERE user_id = ?` |
| `CreateTrack` | Include `user_id` in INSERT |
| `GetTrack` | Add `AND user_id = ?` — prevents accessing another user's track |
| `DeleteTrack` | Add `AND user_id = ?` |

### `backend/db/sessions.go`

Same pattern. Sessions also need `AND user_id = ?` on reads, and `user_id` stored on writes. The session's `user_id` is copied from the track's `user_id` at creation time (already resolved via `GetTrack`).

### `backend/api/tracks.go` and `api/sessions.go`

Extract user from context at the top of each handler:

```go
user := UserFromContext(r.Context())
```

Pass `user.ID` down to every DB call.

---

## Step 7 — Frontend: auth store and API client

### `frontend/src/lib/stores/authStore.ts`

```ts
export type AuthUser = {
    id: number;
    name: string;
    email: string;
    avatarUrl: string;
} | null;

export const currentUser = writable<AuthUser>(null);
export const authLoading = writable(true);
```

### `frontend/src/lib/api/auth.ts`

```ts
fetchCurrentUser(): Promise<AuthUser>          // GET /api/v1/auth/me
logout(): Promise<void>                         // POST /api/v1/auth/logout
loginUrl(provider: 'google' | 'github'): string // returns /api/v1/auth/{provider}/login
```

---

## Step 8 — Frontend: route guard in root layout

### `frontend/src/routes/+layout.ts`

Universal `load` function:
- Calls `GET /api/v1/auth/me`
- If 401 and the current route is not `/login`: throw `redirect(303, '/login')`
- Returns `{ user }` to populate the layout

### `frontend/src/routes/+layout.svelte`

- Receives `user` from `data`, writes it to `authStore`
- All child routes are already behind the guard

---

## Step 9 — Frontend: new routes

### `frontend/src/routes/login/+page.svelte`

- Standalone page (no main layout wrapper, no auth guard)
- Two large buttons: "Continue with Google" and "Continue with GitHub"
- Each is a plain `<a href="/api/v1/auth/{provider}/login">` — full page navigation, not `fetch`, so the OAuth redirect chain works correctly
- Show error message if `?error=` query param is present (set by Go on callback failure)

### `frontend/src/routes/auth/callback/+page.svelte`

- Loading spinner only
- The actual callback is handled server-side by Go; Go redirects to `/` on success or `/login?error=...` on failure
- This route just gives the user something to look at during the redirect

---

## Step 10 — Frontend: header update

### `frontend/src/lib/components/layout/Header.svelte`

- If `$currentUser` is set: display avatar (or initials fallback) + display name + "Sign out" button
  - Sign out: calls `logout()` from `auth.ts`, then `goto('/login')`
- If `$currentUser` is null (and not loading): display "Sign in" link to `/login`

---

## New files summary

```
backend/
├── auth/
│   ├── oauth.go
│   ├── session.go
│   └── state.go
├── models/
│   └── user.go
├── db/
│   ├── users.go
│   └── migrations/
│       ├── sqlite/
│       │   ├── 007_add_users.sql
│       │   └── 008_add_user_id.sql
│       └── postgres/
│           ├── 007_add_users.sql
│           └── 008_add_user_id.sql
└── api/
    └── auth.go

frontend/src/
├── lib/
│   ├── api/
│   │   └── auth.ts
│   └── stores/
│       └── authStore.ts
└── routes/
    ├── login/
    │   └── +page.svelte
    └── auth/callback/
        └── +page.svelte
```

Modified files: `config/config.go`, `api/router.go`, `api/middleware.go`, `api/tracks.go`, `api/sessions.go`, `db/tracks.go`, `db/sessions.go`, `routes/+layout.ts`, `routes/+layout.svelte`, `components/layout/Header.svelte`.

---

## New dependency

```
golang.org/x/oauth2   — OAuth2 client (Google + GitHub sub-packages included)
```

No new frontend dependencies.

---

## Security checklist

| Concern | Mitigation |
|---|---|
| CSRF on OAuth initiation | Random `state` param generated per login, verified in callback via short-lived cookie |
| Session fixation | New random token generated on every login |
| XSS session theft | HttpOnly cookie — JS cannot read the session token |
| CSRF on logout | Logout uses POST; SameSite=Lax blocks cross-site form POSTs |
| Insecure cookie on HTTP | `Secure` flag set in production; relaxed via `DEV_MODE=true` for localhost |
| Cross-user data access | All DB queries filter by `user_id` from authenticated context, never from URL params |
| Open redirect after login | Callback handler only redirects to same-origin paths |

---

## Suggested implementation order

1. Config changes (Step 1)
2. DB migrations 007 + 008 (Step 2)
3. `models/user.go` + `db/users.go` (Step 3)
4. `auth/` package (Step 3)
5. `api/auth.go` + route mounting (Step 4)
6. `RequireAuth` middleware (Step 5)
7. Update `db/tracks.go`, `db/sessions.go`, `api/tracks.go`, `api/sessions.go` (Step 6)
8. Test backend end-to-end with curl before touching the frontend
9. Frontend auth store + API client (Step 7)
10. Root layout guard (Step 8)
11. Login page + callback page (Step 9)
12. Header update (Step 10)
