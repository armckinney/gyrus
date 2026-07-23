# Stage 1: Build Gyrus binary
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/gyrus ./cmd/gyrus

# Stage 2: Minimal runtime image
FROM alpine:3.20

RUN apk add --no-ca-certificates ca-certificates bash curl

COPY --from=builder /bin/gyrus /usr/local/bin/gyrus

WORKDIR /workspace

ENTRYPOINT ["gyrus"]
CMD ["mcp", "serve"]
