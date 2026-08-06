.PHONY: all build_macos_amd build_macos_arm build_linux build clean docker help

# application name
BINARY_FILE=mock

# release
RELEASE=release

# 版本号 (从 cmd/root.go 自动提取)
VERSION := $(shell grep -o '"[0-9]*\.[0-9]*\.[0-9]*"' cmd/root.go | tr -d '"')

all: help

# ==================== 构建命令 ====================

## build_macos_amd: 构建 MacOS AMD64 平台
build_macos_amd:
	@echo "构建 MacOS AMD64 平台..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -gcflags=all=-trimpath=$(GOPATH) -asmflags=all=-trimpath=$(GOPATH) -ldflags "-w -s" -o ${RELEASE}/macos_amd/${BINARY_FILE}
	@echo "Done."

## build_macos_arm: 构建 MacOS ARM64 平台
build_macos_arm:
	@echo "构建 MacOS ARM64 平台..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -gcflags=all=-trimpath=$(GOPATH) -asmflags=all=-trimpath=$(GOPATH) -ldflags "-w -s" -o ${RELEASE}/macos_arm/${BINARY_FILE}
	@echo "Done."

## build_linux: 构建 Linux AMD64 平台
build_linux:
	@echo "构建 Linux AMD64 平台..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -gcflags=all=-trimpath=$(GOPATH) -asmflags=all=-trimpath=$(GOPATH) -ldflags "-w -s" -o ${RELEASE}/linux/${BINARY_FILE}
	@echo "Done."

## build: 构建所有平台
build: clean build_macos_amd build_macos_arm build_linux

# ==================== Docker 命令 ====================

## docker: 构建 Docker 镜像 (二阶段构建)
docker:
	@ARCH=$$(docker version --format '{{.Server.Arch}}'); \
	echo "Docker 宿主机架构: $$ARCH"; \
	if [ "$$ARCH" = "arm64" ]; then \
		PLATFORM=linux/arm64; \
	else \
		PLATFORM=linux/amd64; \
	fi; \
	echo "构建 Docker 镜像 ($$PLATFORM, 版本 ${VERSION})..."; \
	docker build --platform $$PLATFORM -t mock:${VERSION} .
	@echo "Done."

# ==================== 其他命令 ====================

## clean: 清理构建产物
clean:
	@go clean
	@rm -rf ${RELEASE}

## help: 帮助
help:
	@echo "Usage: make <command>"
	@echo ""
	@echo "构建命令:"
	@echo "  make build_macos_amd  构建 MacOS AMD64"
	@echo "  make build_macos_arm  构建 MacOS ARM64"
	@echo "  make build_linux      构建 Linux AMD64"
	@echo "  make build        构建所有平台"
	@echo ""
	@echo "Docker 命令:"
	@echo "  make docker      构建 Docker 镜像 (linux/amd64)"
	@echo ""
	@echo "清理:"
	@echo "  make clean             清理构建产物"
