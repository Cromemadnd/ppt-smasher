# Build stage
FROM node:20-alpine AS build-stage

WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install

COPY . .
RUN pnpm build

# Production stage
FROM nginx:stable-alpine

COPY --from=build-stage /app/dist /usr/share/nginx/html
# 稍后我们将创建 nginx.conf 并挂载或复制进来
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
