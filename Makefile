service ?= 
service_type ?= service

.PHONY: serve start stop logs docs restart golib golib-clean

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
	@docker-compose logs -fn 50 $(service)

watch-dashboard:
	@make -C services/dashboard watch

watch-tenants:
	@make -C services/tenants watch

api:
	-@docker run --rm -p 8080:8080 --init -v $(CURDIR):/project redocly/cli lint /project/tools/openapi/api.yaml
	@echo "Starting live openapi docs"
	-@docker run --rm -p 8080:8080 --init -v $(CURDIR):/project redocly/cli -h 0.0.0.0 preview-docs /project/tools/openapi/api.yaml
	@echo "Stopped live openapi docs"

docs:
	@echo "Starting live docs"
	-@docker run --rm -p 8000:8000 --init -v $(CURDIR):/docs ghcr.io/sensorbucket/mkdocs:latest
	@echo "Stopped live docs"

lint:
	@echo "Running linters..."
	docker run --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:v1.55.2 \
		golangci-lint run --out-format github-actions

python:
ifeq ($(strip $(outdir)),)
	@echo "Error: please specify out location by providing the 'outdir' variable"
else
	@echo "Generating python client from spec"
	@mkdir -p $(outdir)
	@docker run --rm -v $(CURDIR):/sensorbucket -v $(outdir):/target --user `id -u` \
		openapitools/openapi-generator-cli:latest generate -i /sensorbucket/tools/openapi/api.yaml \
		-g python -t /sensorbucket/tools/openapi-templates/python -o /target \
		--git-user-id=sensorbucket.nl --git-repo-id=PythonClient \
		--additional-properties=packageName=sensorbucket,packageUrl='https://sensorbucket.nl'
endif

golib-clean:
ifeq ($(wildcard pkg/api/.openapi-generator/FILES),)
	@echo Nothing to clean 
else
	cat pkg/api/.openapi-generator/FILES | xargs -I_ rm pkg/api/_
	rm pkg/api/.openapi-generator/FILES
endif

golib: golib-clean
	@docker run --rm -v $(CURDIR):/sensorbucket --user `id -u` \
		openapitools/openapi-generator-cli:v6.2.1 generate -i /sensorbucket/tools/openapi/api.yaml \
		-g go -o /sensorbucket/pkg/api \
		--git-host=sensorbucket.nl --git-repo-id=api \
		--enable-post-process-file \
		--additional-properties=packageName=api,packageUrl='https://sensorbucket.nl'

admin: 
	echo '{"schema_id":"default", "traits": {"email":"a@pollex.nl"}}' | http post 127.0.0.1:4434/admin/identities | jq .id | \
	 xargs -I uid echo '{"identity_id":"uid"}' | http post 127.0.0.1:4434/admin/recovery/code
