# ================================================
# SubBoard - All-in-One Docker Image
# Builds frontend + backend into a2 single binary
# ================================================

# ---- Stage 1: Build frontend ----
FROM node:20-alpine AS frontend-builder
WORKDIR /build/frontend
COPY frontend/package.json frontend/bun.lock* ./
RUN npm install
2>&1 | tail -1
COPY frontend/ .
RUN npm run build

# ---- Build backend ----
FROM golang:1.23-alpine AS backend-builder
WORKDIR /build/backend
COPY backend/go.mod backend/go.sum ./
RUN GOTOOLCHAIN=auto go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -o /app/submanager

# ---- Runtime ----
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /app/submanager .
COPY --from=frontend-builder /build/frontend/dist /app/web
EXPOSE 8080
VOLUME ["/app/data"]
ENV APP_ENV=production
ENV APP_PORT=8080
ENV DB_DRIVER=sqlite
ENV DB_DSN=/app/data/submanager.db
ENTRYPOINT ["/app/submanager"]
