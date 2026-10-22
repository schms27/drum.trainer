# Branding

## Logo

Two crossed drumsticks above four beat lights — the first, accented beat in
indigo, like the app's metronome. Drawn in VectorCraft.

| File | Use |
|---|---|
| `frontend/static/logo.svg` | Header, login page (256×256, rounded badge) |
| `frontend/static/favicon.svg` | Browser tab — bolder sticks, no beat lights, so it stays legible at 16 px |
| `frontend/static/apple-touch-icon.png` | iOS home screen (180×180, square — iOS rounds the corners itself) |
| `frontend/static/icons/icon-192.png`, `icon-512.png` | PWA install icons — `logo.svg` rendered at 192 and 512 px |
| `frontend/static/icons/icon-maskable-512.png` | PWA maskable icon (Android adaptive icons) — rendered from `docs/branding/icon-maskable.svg` |
| `docs/branding/icon-maskable.svg` | Logo artwork scaled to 85 % on a full-bleed `#3a3a42` square, so it stays inside the maskable safe zone (centre circle, 40 % radius) |
| `docs/branding/logo-source.vectorcraft` | Editable source with all artboards (including the earlier drafts) |
| `docs/branding/stick.py` | Generates the drumstick outlines (butt, shaft, taper, neck, acorn tip) as SVG path data |

Artboards in the source file: **B1c Sticks #3a3a42** is the logo, **B Favicon (bold)**
the favicon and **Touch icon (square)** the Apple touch icon. Export them as SVG
(PNG at 180 px for the touch icon) and replace the `<title>` with "Drum Trainer".

## Colours

| Role | Value |
|---|---|
| Badge | `#3a3a42` — a step lighter than the dark header (`#18181b`) so it stays visible on it |
| Sticks | `#f3cda5` (front), `#e2ad80` (back), highlights `#fde6cf` / `#efc9a3` |
| Accented beat | `#818cf8` (the app's dark-mode accent text colour) |
| Other beats | `#8b8b94` |

## Typeface

**Outfit** (weight 700, SIL Open Font License) for the wordmark and page titles,
bundled with the app via `@fontsource/outfit` — no requests to font servers.
Registered as the `font-brand` Tailwind utility in `frontend/src/app.css`.
Body text keeps the system font stack.
