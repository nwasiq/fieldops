# Build stage: static binary, no cgo (the SQLite driver is pure Go and only
# used by tests; the runtime driver is pgx).
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/fieldops .

# Runtime stage: distroless static, non-root. Zoneinfo is embedded in the
# binary (time/tzdata), so no OS packages are needed for Europe/London.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/fieldops /fieldops
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/fieldops"]
