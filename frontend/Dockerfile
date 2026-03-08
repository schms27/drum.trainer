# ── Build stage ──────────────────────────────────────────────────────────────
FROM node:22-alpine AS builder

WORKDIR /build

# Cache npm install separately from source changes.
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .
RUN npm run build

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM node:22-alpine

WORKDIR /app

# adapter-node build output is self-contained but still resolves imports from
# node_modules at runtime, so we install only production dependencies.
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --omit=dev

COPY --from=builder /build/build ./build

ENV NODE_ENV=production \
    PORT=3000

EXPOSE 3000
CMD ["node", "build/index.js"]
