# ---------- 1. 前端构建 ----------
FROM node:22-alpine AS webbuild
WORKDIR /web
COPY web/package.json web/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm build

# ---------- 2. Go 构建 ----------
FROM golang:1.27-alpine AS gobuild
WORKDIR /src
COPY server/ .
COPY --from=webbuild /web/dist ./internal/frontend/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /studypilot ./cmd/server

# ---------- 3. 运行时 ----------
FROM alpine:3.21
WORKDIR /app
COPY --from=gobuild /studypilot .
EXPOSE 8080
VOLUME ["/app/data"]
ENTRYPOINT ["./studypilot"]