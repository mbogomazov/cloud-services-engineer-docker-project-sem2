# syntax=docker/dockerfile:1

# ---- build stage -------------------------------------------------------------
ARG GO_VERSION=1.17
FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

# Cache module downloads: copy manifests first so this layer is reused
# until go.mod / go.sum actually change.
COPY go.mod go.sum ./
RUN go mod download

# Build the statically-linked binaries.
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-w -s" -o /out/app ./cmd/api && \
    go build -trimpath -ldflags="-w -s" -o /out/healthcheck ./cmd/healthcheck

# ---- final stage -------------------------------------------------------------
# distroless/static:nonroot -> no shell, no package manager, runs as uid 65532.
FROM gcr.io/distroless/static:nonroot AS final

WORKDIR /

COPY --from=build /out/app /app
COPY --from=build /out/healthcheck /healthcheck

USER nonroot:nonroot

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/healthcheck"]

ENTRYPOINT ["/app"]
