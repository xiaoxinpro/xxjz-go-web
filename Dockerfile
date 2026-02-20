# Stage 1: build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci 2>/dev/null || npm install
COPY frontend/ .
RUN npm run build

# Stage 2: build backend
FROM golang:1.21-alpine AS backend
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=1 go build -o /server ./cmd/server

# Stage 3: runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend /server .
COPY --from=frontend /app/frontend/dist ./static
COPY backend/migrations ./migrations
COPY config.yaml .
ENV CONFIG=config.yaml
EXPOSE 8080
VOLUME /app/data
CMD ["./server"]
