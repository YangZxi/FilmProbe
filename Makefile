# Makefile

# 定义变量
IMAGE_NAME = yangzxi/film-probe
TAG = latest
DOCKERFILE_PATH = Dockerfile
BUILD_DIR = .

# 默认目标
all: build push

# 构建 Docker 镜像
build:
	docker buildx build --platform linux/amd64 -t $(IMAGE_NAME):$(TAG) -f $(DOCKERFILE_PATH) $(BUILD_DIR)

# 推送 Docker 镜像到 Docker Hub
push:
	docker push $(IMAGE_NAME):$(TAG)

# 清理未使用的镜像
clean:
	docker rmi $(IMAGE_NAME):$(TAG) || true

# 显示帮助信息
help:
	@echo "Usage:"
	@echo "  make build  # Build the Docker image"
	@echo "  make push   # Push the Docker image to Docker Hub"
	@echo "  make clean  # Remove the Docker image"
	@echo "  make all    # Build and push the Docker image"
