# Stage 1: Build
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/fetcher .

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/fetcher /fetcher

USER nonroot:nonroot

ENTRYPOINT ["/fetcher"]
