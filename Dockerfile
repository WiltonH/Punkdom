# syntax=docker/dockerfile:1

FROM node:22-alpine AS web-builder
WORKDIR /src/web
RUN corepack enable
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web ./
RUN pnpm build

FROM golang:1.26.1-alpine AS go-builder
WORKDIR /src
ARG PUNKDOM_VERSION=dev
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /src/web/dist ./web/dist
RUN go build -trimpath -ldflags "-s -w -X punkdom/internal/buildinfo.Version=${PUNKDOM_VERSION}" -o /out/punkdom ./cmd/punkdom

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
  && addgroup -S punkdom \
  && adduser -S -G punkdom -h /data punkdom \
  && mkdir -p /app /data \
  && chown -R punkdom:punkdom /data
WORKDIR /data
COPY --from=go-builder /out/punkdom /app/punkdom
COPY --from=web-builder /src/web/dist /app/web
COPY skills /app/skills
COPY README.md README.en.md CHANGELOG.md LICENSE NOTICE /app/
ENV PUNKDOM_DOCKER=1 \
    PUNKDOM_DIR=/data/.punkdom \
    PUNKDOM_WORKSPACE=/data/workspaces \
    PUNKDOM_WEB_DIR=/app/web \
    PUNKDOM_SKILLS_DIR=/app/skills \
    PUNKDOM_BACKEND_PORT=8080
EXPOSE 8080
USER punkdom
CMD ["/app/punkdom", "--no-open", "--port", "8080"]
