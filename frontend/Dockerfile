# syntax=docker/dockerfile:1

# ---- build stage -------------------------------------------------------------
# Node 16 is required: vue-cli 4.5 / webpack 4 break on Node 17+ OpenSSL.
ARG NODE_VERSION=16
FROM node:${NODE_VERSION}-alpine AS build

WORKDIR /app

# Cache npm install: copy lockfiles first.
COPY package.json package-lock.json ./
RUN npm ci

# Build-time configuration (see README "Configuration").
#  - VUE_APP_API_URL: axios base URL. Empty => relative paths, served through
#    the nginx reverse proxy in this image.
#  - PUBLIC_PATH: asset base path (vue.config.js), '/' when served at root.
ARG VUE_APP_API_URL=""
ARG PUBLIC_PATH="/"
ENV VUE_APP_API_URL=${VUE_APP_API_URL} \
    PUBLIC_PATH=${PUBLIC_PATH} \
    NODE_ENV=production

COPY . .
RUN npm run build

# ---- final stage -------------------------------------------------------------
# Unprivileged nginx: runs as uid 101, listens on 8080 (non-root cannot bind 80).
# docker-compose publishes host :80 -> container :8080.
FROM nginxinc/nginx-unprivileged:1.27-alpine AS final

# Replace the default site with our SPA + reverse-proxy config.
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Static assets from the build stage.
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/ >/dev/null 2>&1 || exit 1
