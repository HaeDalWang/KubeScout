# Stage 1: Build Frontend
FROM --platform=$BUILDPLATFORM node:20-alpine AS frontend-builder
WORKDIR /app/web

# Install dependencies
COPY web/package.json web/package-lock.json ./
RUN npm ci

# Copy source and build
COPY web/ .
RUN npm run build

# Stage 2: Build Backend (Cross-Compilation)
# --platform=$BUILDPLATFORM: 항상 네이티브 플랫폼에서 실행
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS backend-builder

# TARGETPLATFORM, TARGETOS, TARGETARCH는 Docker Buildx가 자동 주입
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# frontend-builder 스테이지에서 빌드된 결과물 복사
COPY --from=frontend-builder /app/web/dist ./internal/ui/dist

# Go Cross-Compilation (네이티브에서 타겟 아키텍처용 바이너리 빌드)
# QEMU 에뮬레이션 없이 빠르게 빌드 가능
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-w -s" -o kubescout ./cmd/kubescout

# Stage 3: Final Image (타겟 플랫폼용)
FROM gcr.io/distroless/static:nonroot
WORKDIR /

COPY --from=backend-builder /app/kubescout /kubescout

EXPOSE 8080

USER 65532:65532

ENTRYPOINT ["/kubescout"]
