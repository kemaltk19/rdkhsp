FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
# Build sırasında API adresi gömülür (compose'dan build arg ile gelir).
ARG VITE_API_URL=http://localhost:9001/api
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

FROM nginx:alpine
# SPA fallback için özel nginx config (history-mode router 404 olmasın).
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
