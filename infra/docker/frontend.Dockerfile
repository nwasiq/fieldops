# Builds the frontend bundle and serves it with nginx on port 3000, proxying /api to the
# backend service. Used only by docker-compose.yml (profile "app"); the deployed frontend is
# the same `npm run build` output uploaded to S3 behind CloudFront (infra/modules/edge).
# Build context is ./frontend; this file lives outside it so frontend/ stays the sibling's.
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
