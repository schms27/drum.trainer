.PHONY: dev backend frontend build clean test

## Run both backend and frontend in parallel
dev:
	@$(MAKE) -j2 backend frontend

## Run Go backend (http://localhost:8080)
backend:
	cd backend && go run ./main.go

## Run SvelteKit frontend dev server (http://localhost:5173)
frontend:
	cd frontend && npm run dev

## Production build (frontend output → frontend/build/, then go build)
build:
	cd frontend && npm run build
	cd backend && go build -o drum-trainer .

## Run all tests (Go + frontend)
test:
	cd backend && go test ./...
	cd frontend && npm test

## Remove build artifacts
clean:
	rm -rf frontend/build frontend/.svelte-kit backend/drum-trainer backend/data
