# Development environment commands
dev-up:
	docker compose -f docker-compose.yml up -d

dev-down:
	docker compose -f docker-compose.yml down

dev-down-volumes:
	docker compose -f docker-compose.yml down -v

dev-logs:
	docker compose -f docker-compose.yml logs -f

# Helper commands
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development Environment:"
	@echo "  dev-up                    - Start development environment in detached mode"
	@echo "  dev-down                  - Stop development environment"
	@echo "  dev-down-volumes          - Stop development environment and remove volumes"
	@echo "  dev-logs                  - View logs from all services"
	@echo ""
	@echo "Manual Testing:"
	@echo "  See docs/manual_testing_guide.md for comprehensive testing scenarios"
	@echo ""
	@echo "Monitoring:"
	@echo "  Temporal UI:    http://localhost:8080"
	@echo "  Grafana:        http://localhost:3001 (admin/admin)"
	@echo "  Prometheus:     http://localhost:9090"

.PHONY: dev-up dev-down dev-down-volumes dev-logs help
