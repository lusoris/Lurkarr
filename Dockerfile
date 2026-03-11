# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.26.1-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/build ./frontend/build
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o lurkarr ./cmd/lurkarr

# Stage 3: Runtime (scratch + certs for minimal attack surface)
FROM alpine:3.21 AS certs
RUN apk add --no-cache ca-certificates tzdata

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=certs /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=backend /app/lurkarr /lurkarr
EXPOSE 9705
ENTRYPOINT ["/lurkarr"]
