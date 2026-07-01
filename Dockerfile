# Stage 1: Build Go backend
FROM golang:1.22-alpine AS go-builder
WORKDIR /app/backend
COPY backend/ .
RUN go mod download && go build -o /server ./cmd/server/

# Stage 2: Build Next.js frontend
FROM node:20-alpine AS next-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Stage 3: Production image
FROM alpine:3.19
RUN apk add --no-cache nginx

COPY --from=go-builder /server /server
COPY --from=next-builder /app/frontend/out /var/www/html
COPY nginx/nginx.conf /etc/nginx/nginx.conf

RUN adduser -D -h /home/appuser appuser
USER appuser

EXPOSE 7860
CMD ["sh", "-c", "/server & nginx -g 'daemon off;'"]