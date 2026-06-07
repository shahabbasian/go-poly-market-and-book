# Stage 1: Build
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Copy source and module files, then resolve dependencies
COPY . .
RUN go mod tidy && go mod download

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/fetcher .

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/fetcher /fetcher

USER nonroot:nonroot

ENTRYPOINT ["/fetcher"]
