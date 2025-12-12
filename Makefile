# 项目配置
APP_NAME := weworkapi
BINARY_NAME := $(APP_NAME)
MAIN_FILE := sample.go
PORT := 8090
PID_FILE := $(APP_NAME).pid
LOG_FILE := $(APP_NAME).log

# Go 编译参数
GO := go
GO_BUILD := $(GO) build
GO_RUN := $(GO) run

.PHONY: build start stop restart clean help

# 编译项目
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO_BUILD) -o $(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BINARY_NAME)"

# 启动服务（后台运行）
start: build
	@if [ -f $(PID_FILE) ]; then \
		PID=$$(cat $(PID_FILE)); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "Service is already running (PID: $$PID)"; \
			exit 1; \
		else \
			rm -f $(PID_FILE); \
		fi; \
	fi
	@echo "Starting $(BINARY_NAME) on port $(PORT)..."
	@nohup ./$(BINARY_NAME) > $(LOG_FILE) 2>&1 & \
		echo $$! > $(PID_FILE)
	@sleep 1
	@if [ -f $(PID_FILE) ]; then \
		PID=$$(cat $(PID_FILE)); \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "Service started successfully (PID: $$PID)"; \
			echo "Log file: $(LOG_FILE)"; \
		else \
			echo "Service failed to start. Check $(LOG_FILE) for details."; \
			rm -f $(PID_FILE); \
			exit 1; \
		fi; \
	fi

# 停止服务
stop:
	@if [ ! -f $(PID_FILE) ]; then \
		echo "PID file not found. Service may not be running."; \
		exit 1; \
	fi
	@PID=$$(cat $(PID_FILE)); \
	if ps -p $$PID > /dev/null 2>&1; then \
		echo "Stopping $(BINARY_NAME) (PID: $$PID)..."; \
		kill $$PID; \
		sleep 1; \
		if ps -p $$PID > /dev/null 2>&1; then \
			echo "Force killing process..."; \
			kill -9 $$PID; \
		fi; \
		rm -f $(PID_FILE); \
		echo "Service stopped."; \
	else \
		echo "Process not found. Removing stale PID file."; \
		rm -f $(PID_FILE); \
	fi

# 重启服务
restart: stop start
	@echo "Service restarted."

# 清理编译文件和日志
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME) $(PID_FILE) $(LOG_FILE)
	@echo "Clean complete."

# 显示帮助信息
help:
	@echo "Available targets:"
	@echo "  make build    - Build the application"
	@echo "  make start    - Start the service in background"
	@echo "  make stop     - Stop the running service"
	@echo "  make restart  - Restart the service"
	@echo "  make clean    - Remove build artifacts and logs"
	@echo "  make help     - Show this help message"
	@echo ""
	@echo "Configuration:"
	@echo "  APP_NAME:     $(APP_NAME)"
	@echo "  PORT:         $(PORT)"
	@echo "  PID_FILE:     $(PID_FILE)"
	@echo "  LOG_FILE:     $(LOG_FILE)"

