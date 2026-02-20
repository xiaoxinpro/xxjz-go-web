.PHONY: build docker-build docker-push test frontend backend clean

# 构建输出目录（前后端产物均在此，可直接运行）
BUILD_DIR := build

# 若 build 已存在则删除并重建，然后构建前后端并复制运行所需文件
build:
	rm -rf $(BUILD_DIR) && mkdir -p $(BUILD_DIR)
	cd frontend && (npm ci 2>/dev/null || npm install) && npm run build
	cd backend && go build -o ../$(BUILD_DIR)/server ./cmd/server
	cp config.yaml $(BUILD_DIR)/
	mkdir -p $(BUILD_DIR)/static && cp -r frontend/dist/. $(BUILD_DIR)/static/
	cp -r backend/migrations $(BUILD_DIR)/
	@printf '小歆记账 - 构建输出目录\n运行方式: ./server\n默认端口 8080，数据目录 ./data，配置 config.yaml\n' > $(BUILD_DIR)/运行说明.txt
	@echo "Build done. Run: cd $(BUILD_DIR) && ./server"

# 仅构建前端（输出到 frontend/dist，用于开发）
frontend:
	cd frontend && (npm ci 2>/dev/null || npm install) && npm run build

# 仅构建后端到 build/server（需先有 build 目录）
backend:
	mkdir -p $(BUILD_DIR)
	cd backend && go build -o ../$(BUILD_DIR)/server ./cmd/server

# 清理构建目录
clean:
	rm -rf $(BUILD_DIR) frontend/dist

# Docker
docker-build:
	docker build -t xxjz-go-web:latest .

docker-push: docker-build
	docker tag xxjz-go-web:latest $(REGISTRY)/xxjz-go-web:latest
	docker push $(REGISTRY)/xxjz-go-web:latest

# Tests
test:
	cd backend && go test ./...

test-race:
	cd backend && go test -race ./...
