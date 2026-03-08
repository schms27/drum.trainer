# drum.trainer

A web app for drummers to practice along with MIDI tracks and get scored on timing accuracy. Upload a MIDI file, connect your e-drum kit via USB, record a take, and see a detailed breakdown of hit rate, timing offset, and ghost notes per instrument.

## Features

- Upload standard MIDI files as practice targets
- Record live MIDI input from any e-drum kit or controller
- Per-instrument scoring: hit rate, timing offset, consistency, ghost notes
- Pilot tracks — upload audio files (WAV, MP3) that play back in sync with your first hit at any practice tempo
- Adjustable practice tempo with automatic time-scaling
- Session history and weekly practice statistics
- Sign in with Google or GitHub

## Browser Requirements

The recording feature requires the [Web MIDI API](https://developer.mozilla.org/en-US/docs/Web/API/Web_MIDI_API), which is currently supported in Chromium-based browsers (Chrome, Edge, Opera). Firefox and Safari do not support Web MIDI.

---

## Self-Hosting with Docker Compose

This is the recommended way to run drum.trainer in production.

### Prerequisites

- Docker and Docker Compose
- A Google or GitHub OAuth application (see below)
- A reverse proxy terminating TLS (nginx, Caddy, Traefik, Pangolin, etc.)

### 1. Create OAuth credentials

**Google**
1. Go to [Google Cloud Console](https://console.cloud.google.com/) → APIs & Services → Credentials
2. Create an OAuth 2.0 Client ID (Web application)
3. Add `https://your-domain.com/api/v1/auth/google/callback` as an authorised redirect URI

**GitHub**
1. Go to GitHub → Settings → Developer settings → OAuth Apps → New OAuth App
2. Set the callback URL to `https://your-domain.com/api/v1/auth/github/callback`

You can configure one or both providers; unused ones can be left blank.

### 2. Create a `.env` file

```env
# Database
POSTGRES_USER=drum
POSTGRES_PASSWORD=a-strong-random-password

# Public URL of the app (no trailing slash)
BASE_URL=https://your-domain.com

# Must match BASE_URL exactly (used by SvelteKit CSRF protection)
ORIGIN=https://your-domain.com

# Random secret for signing OAuth CSRF state cookies — generate with:
#   openssl rand -hex 32
SESSION_SECRET=replace-with-a-random-secret

# OAuth — configure at least one provider
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=

# Set to true only for local HTTP development (disables Secure cookie flag)
DEV_MODE=false

# Host path where persistent data (uploads, postgres) will be stored
CONTAINER_MOUNTS_PATH=/opt/docker-data
```

### 3. Copy `compose.yml` and `nginx.conf`

Both files are provided in the root of this repository:

- [`compose.yml`](compose.yml) — Docker Compose service definitions
- [`nginx.conf`](nginx.conf) — nginx reverse proxy configuration

Copy them to your server and place `nginx.conf` at the path referenced by the `proxy` volume mount in `compose.yml`.

### 4. Start

```bash
docker compose up -d
```

Point your external reverse proxy at port `9980`. The app will be available at the `BASE_URL` you configured.

---

## Environment Variable Reference

### Backend

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | No | — | PostgreSQL connection string. If unset, SQLite is used (`DB_PATH`). |
| `DB_PATH` | No | `./data/drum_trainer.db` | SQLite database file path (ignored when `DATABASE_URL` is set). |
| `UPLOAD_DIR` | No | `./data/uploads` | Directory for uploaded MIDI and pilot audio files. |
| `PORT` | No | `8080` | Port the backend listens on. |
| `BASE_URL` | Yes | `http://localhost:8080` | Public base URL. Used to build OAuth callback URLs. |
| `SESSION_SECRET` | Yes | *(insecure default)* | Secret for signing OAuth CSRF state cookies. Use `openssl rand -hex 32`. |
| `GOOGLE_CLIENT_ID` | No | — | Google OAuth client ID. |
| `GOOGLE_CLIENT_SECRET` | No | — | Google OAuth client secret. |
| `GITHUB_CLIENT_ID` | No | — | GitHub OAuth client ID. |
| `GITHUB_CLIENT_SECRET` | No | — | GitHub OAuth client secret. |
| `MAX_UPLOAD_SIZE_MB` | No | `16` | Maximum MIDI file upload size in MB. |
| `MAX_PILOT_SIZE_MB` | No | `150` | Maximum pilot audio file upload size in MB. |
| `MAX_SESSION_HITS` | No | `20000` | Maximum number of recorded hits accepted per session submission. |
| `DEV_MODE` | No | `false` | Set to `true` for local HTTP development. Disables the `Secure` flag on session cookies. |

### Frontend

| Variable | Required | Description |
|---|---|---|
| `ORIGIN` | Yes | Must match the URL the browser uses to reach the app. Required for SvelteKit CSRF protection. |

---

## Local Development

### Prerequisites

- Go 1.24.2+
- Node.js 20+
- A Chromium-based browser for MIDI recording

### Setup

```bash
# Backend
cd backend
go run ./main.go

# Frontend (separate terminal)
cd frontend
npm install
npm run dev
```

The backend runs on `http://localhost:8080` and the frontend on `http://localhost:5173`. The Vite dev server proxies `/api/*` to the backend, so no CORS configuration is needed locally.

For OAuth to work locally you will need to set the backend environment variables and set `DEV_MODE=true`. Alternatively, skip login entirely by temporarily disabling the `RequireAuth` middleware during development.

### Running Tests

```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
npm test
```

---

## Architecture

| Layer | Technology |
|---|---|
| Frontend | SvelteKit 2 + Svelte 5, Tailwind CSS 4, D3 |
| Backend | Go 1.24, chi router |
| Database | PostgreSQL (production) or SQLite (development) |
| Auth | OAuth 2.0 via Google / GitHub, server-side session cookie |
| MIDI parsing | gitlab.com/gomidi/midi/v2 |
| Scoring | Greedy nearest-neighbour, ±200ms window, per instrument group |

Migrations run automatically on startup. Both SQLite and PostgreSQL migration sets are embedded in the binary.

## Releasing a New Version

Releases are driven by git tags. Pushing a tag triggers the CI pipeline, which runs all tests and publishes versioned Docker images to the Codeberg registry.

```bash
git tag v1.2.3
git push origin v1.2.3
```

This produces the following image tags for both `drum-trainer-backend` and `drum-trainer-frontend`:

| Tag | Example |
|---|---|
| Exact version | `1.2.3` |
| Major.minor | `1.2` |
| Major | `1` |
| Short SHA | `a3f9c12` |

`latest` is only updated on pushes to `main`, not on tag pushes.

To deploy the new version, update the image tags in your `compose.yml` (or pull `latest` if you track that) and run:

```bash
docker compose pull
docker compose up -d
```

## License

MIT
