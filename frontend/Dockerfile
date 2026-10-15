# ── Build stage ──────────────────────────────────────────────────────────────
FROM node:22-alpine AS builder

ARG APP_VERSION=dev

WORKDIR /build

# Cache npm install separately from source changes.
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .
RUN VITE_APP_VERSION=$APP_VERSION npm run build

# ── Runtime stage ─────────────────────────────────────────────────────────────
# The app is a static SPA, so plain nginx serves it — no Node runtime needed.
FROM nginx:alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /build/build /usr/share/nginx/html

EXPOSE 3000
