# Stage 1: Build the frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /app/client

COPY client/package.json client/package-lock.json ./
RUN npm ci

COPY client/ .
RUN npm run build

# Stage 2: Build the binary
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy built frontend assets
COPY --from=frontend-builder /app/internal/adapters/http/static/dist ./internal/adapters/http/static/dist

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o mailraven ./cmd/mailraven
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o mailraven-cli ./cmd/mailraven-cli

# Stage 3: Production image
FROM gcr.io/distroless/static-debian12 AS production

WORKDIR /app

COPY --from=builder /app/mailraven /app/mailraven
COPY --from=builder /app/mailraven-cli /app/mailraven-cli
COPY --from=builder /app/deployment/config.example.yaml /etc/mailraven/config.yaml

EXPOSE 25 8080

VOLUME ["/data"]

ENTRYPOINT ["/app/mailraven", "serve", "--config", "/etc/mailraven/config.yaml"]
