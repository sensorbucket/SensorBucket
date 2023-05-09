service ?= 
service_type ?= service

.PHONY: serve start stop logs docs restart

serve:
	@echo "Watching service: $(service)"
	@reflex -r '\.go$$' -s -t 500ms -- go run $(service_type)s/$(service)/main.go

start:
	@echo "Starting development environment..."
	@docker-compose -f $(CURDIR)/docker-compose.yaml up -d $(service)
	@echo "Development environment running"

stop:
	@echo "Stopping development environment..."
	@docker-compose -f $(CURDIR)/docker-compose.yaml stop $(service)
	@echo "Development environment stopped"

restart:
	@docker-compose -f $(CURDIR)/docker-compose.yaml restart $(service)

logs:
	@docker-compose logs -f $(service)

api:
	@echo "Starting live openapi docs"
	-@docker run --rm -p 8080:8080 --init -v $(CURDIR):/project redocly/cli -h 0.0.0.0 preview-docs /project/tools/openapi/api.yaml
	@echo "Stopped live openapi docs"

docs:
	@echo "Starting live docs"
	-@docker run --rm -p 8080:8000 --init -v $(CURDIR):/docs ghcr.io/sensorbucket/mkdocs:latest
	@echo "Stopped live docs"
