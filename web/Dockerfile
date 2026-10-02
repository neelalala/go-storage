FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

COPY . ./
RUN npm run build

FROM nginx:1.31-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

ENV GATEWAY_URL=http://gateway:80

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
