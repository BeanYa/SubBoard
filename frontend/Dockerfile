# Build stage
FROM oven/bun:1-alpine AS builder
WORKDIR /build
COPY package*.json ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run build

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates nginx
WORKDIR /dist
COPY --from=builder /build/dist .
# Nginx runs via docker-compose, not in container
CMD ["echo", "Frontend built. Serve via nginx."]