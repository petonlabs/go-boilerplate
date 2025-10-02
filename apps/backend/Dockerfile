# syntax=docker/dockerfile:1.7

FROM golang:trixie AS deps
WORKDIR /app
RUN apk add --no-cache ca-certificates git build-base
COPY apps/backend/go.mod apps/backend/go.sum ./ 
RUN go mod download

FROM deps AS build
WORKDIR /app
COPY apps/backend .
# Leverage build cache
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/bin/server ./cmd/go-boilerplate

# Distroless for minimal attack surface
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /app/bin/server /app/server
COPY --from=build /app/static /app/static
COPY --from=build /app/templates /app/templates
COPY --from=deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD ["/app/server", "-healthcheck"]
ENTRYPOINT ["/app/server"]
