# ---- Stage 1: Build ------------------------------------------------
FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm install --prefer-offline

COPY . .
RUN npm run build -- --configuration production

# ---- Stage 2: Serve ------------------------------------------------
FROM nginxinc/nginx-unprivileged:1.27-alpine AS runtime

COPY --from=builder /app/dist/dashboard/browser /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
