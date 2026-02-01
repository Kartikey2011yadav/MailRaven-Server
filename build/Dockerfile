# Stage 1: Build the frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app/client

# Copy dependencies first for caching
COPY client/package.json client/package-lock.json ./
RUN npm ci

# Copy source and build
COPY client/ .
RUN npm run build

# Stage 2: Build the binary
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Copy built frontend assets
COPY --from=frontend-builder /app/client/dist ./internal/adapters/http/static/dist

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o mailraven ./cmd/mailraven
# Also build CLI
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o mailraven-cli ./cmd/mailraven-cli

# Stage 2: Create the final image
FROM gcr.io/distroless/static-debian12 AS production

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/mailraven /app/mailraven
COPY --from=builder /app/mailraven-cli /app/mailraven-cli

# Copy config example (though usually mounted)
COPY --from=builder /app/deployment/config.example.yaml /etc/mailraven/config.yaml

# Expose ports (SMTP, API)
EXPOSE 25 8080

# Volume for persistent data
VOLUME ["/data"]

ENTRYPOINT ["/app/mailraven", "serve", "--config", "/etc/mailraven/config.yaml"]
