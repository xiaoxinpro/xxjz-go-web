.PHONY: build docker-build docker-push test frontend backend

# Build frontend (production)
frontend:
	cd frontend && npm ci 2>/dev/null || npm install && npm run build

# Build backend binary (from repo root, config at ./config.yaml)
backend:
	cd backend && go build -o ../server ./cmd/server

# Build both
build: frontend backend

# Docker
docker-build:
	docker build -t xxjz-go-web:latest .

docker-push: docker-build
	docker tag xxjz-go-web:latest $(REGISTRY)/xxjz-go-web:latest
	docker push $(REGISTRY)/xxjz-go-web:latest

# Tests
test:
	cd backend && go test ./...

test-race:
	cd backend && go test -race ./...
