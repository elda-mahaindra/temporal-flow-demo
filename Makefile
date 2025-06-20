# Development environment commands
dev-up:
	docker-compose -f docker-compose.yml up -d

dev-down:
	docker-compose -f docker-compose.yml down

dev-down-volumes:
	docker-compose -f docker-compose.yml down -v

# Integration testing commands
test-integration:
	docker-compose -f docker-compose.yml -f integration-tests/docker-compose.test.yml up --build --abort-on-container-exit integration-tests

test-integration-up:
	docker-compose -f docker-compose.yml -f integration-tests/docker-compose.test.yml up -d

test-integration-down:
	docker-compose -f docker-compose.yml -f integration-tests/docker-compose.test.yml down

test-integration-down-volumes:
	docker-compose -f docker-compose.yml -f integration-tests/docker-compose.test.yml down -v

test-integration-logs:
	docker logs integration-tests

# Helper commands
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development Environment:"
	@echo "  dev-up                    - Start development environment in detached mode"
	@echo "  dev-down                  - Stop development environment"
	@echo "  dev-down-volumes          - Stop development environment and remove volumes"
	@echo ""
	@echo "Integration Testing:"
	@echo "  test-integration          - Run integration tests (build, run, exit)"
	@echo "  test-integration-up       - Start integration test environment"
	@echo "  test-integration-down     - Stop integration test environment"
	@echo "  test-integration-down-volumes - Stop integration test environment and remove volumes"
	@echo "  test-integration-logs     - View integration test logs"

.PHONY: dev-up dev-down dev-down-volumes test-integration test-integration-up test-integration-down test-integration-down-volumes test-integration-logs help
