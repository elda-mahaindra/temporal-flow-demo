# Development environment commands
dev-up:
	docker-compose -f docker-compose.yml up -d

dev-down:
	docker-compose -f docker-compose.yml down

dev-down-volumes:
	docker-compose -f docker-compose.yml down -v

# Helper commands
help:
	@echo "Available commands:"
	@echo "  dev-up           - Start development environment in detached mode"
	@echo "  dev-down         - Stop development environment"
	@echo "  dev-down-volumes - Stop development environment and remove volumes"

.PHONY: dev-up dev-down dev-down-volumes help
