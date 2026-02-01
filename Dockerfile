FROM golang:1.24-bullseye AS builder

WORKDIR /src

COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/cash-track ./cmd/server

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libsqlite3-0 \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/cash-track /app/cash-track
COPY web /app/web

ENV SERVER_PORT=8080
ENV DATABASE_URL=/data/cash-track.db
ENV UPLOAD_DIR=/data/uploads

EXPOSE 8080

CMD ["/app/cash-track"]
