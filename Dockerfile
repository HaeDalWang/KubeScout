# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web

# Install dependencies
COPY web/package.json web/package-lock.json ./
RUN npm ci

# Copy source and build
COPY web/ .
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Embed Frontend Assets
# frontend-builder 스테이지에서 빌드된 결과물 복사
COPY --from=frontend-builder /app/web/dist ./internal/ui/dist

# Build Go binary
# CGO_ENABLED=0 for static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o kubescout ./cmd/kubescout

# Stage 3: Final Image
FROM gcr.io/distroless/static:nonroot
WORKDIR /

COPY --from=backend-builder /app/kubescout /kubescout

EXPOSE 8080

# Run as non-root user
USER 65532:65532

ENTRYPOINT ["/kubescout"]
