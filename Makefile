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
	docker pull golangci/golangci-lint:latest
	docker run --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:latest \
		golangci-lint run --out-format colored-line-number

python-clean:
ifeq ($(strip $(OUTDIR)),)
	@echo "Error: please specify out location by providing the 'OUTDIR' variable"
else
ifneq ($(wildcard $(OUTDIR)/.openapi-generator/FILES),)
	cat $(OUTDIR)/.openapi-generator/FILES | xargs -I_ rm $(OUTDIR)/_
	rm $(OUTDIR)/.openapi-generator/FILES
endif
endif

SRC_VERSION ?= $(shell git describe --tags --dirty)
python: python-clean
ifeq ($(strip $(OUTDIR)),)
	@echo "Error: please specify out location by providing the 'OUTDIR' variable"
else
	@echo "Generating python client from spec with version: $(SRC_VERSION)"
	@mkdir -p $(OUTDIR)
	@docker run --rm -v $(CURDIR):/sensorbucket -v $(OUTDIR):/target --user `id -u` \
		openapitools/openapi-generator-cli:latest generate -i /sensorbucket/tools/openapi/api.yaml \
		-g python -t /sensorbucket/tools/openapi-templates/python -o /target \
		--git-user-id=sensorbucket.nl --git-repo-id=PythonClient \
		--additional-properties=packageName=sensorbucket,packageUrl='https://sensorbucket.nl,packageVersion=$(SRC_VERSION)'
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

USER_EMAIL ?= a@pollex.nl
USER_SCHEMA ?= default
usercreate: 
	@echo '{"schema_id":"$(USER_SCHEMA)", "traits": {"email":"$(USER_EMAIL)"}}' | http post 127.0.0.1:4434/admin/identities | jq -r .id

userfind: 
	@http get 127.0.0.1:4434/admin/identities\?credentials_identifier=$(USER_EMAIL) | jq -r .[0].id

USER_ID = 
EXPIRES_IN ?= 5m
userrecover:
	@echo '{"identity_id":"$(USER_ID)","expires_in":"$(EXPIRES_IN)"}' | http post 127.0.0.1:4434/admin/recovery/code

oathkeeper:
	-@mkdir -p $(CURDIR)/tools/oathkeeper
	@docker run --rm --init -v $(CURDIR):/project redocly/cli bundle /project/tools/openapi/api.yaml > $(CURDIR)/tools/oathkeeper/bundled_openapi.yaml
	openkeeper generate --config $(CURDIR)/tools/oathkeeper/openkeeper.toml

.PHONY: webdeps
webdeps: oathkeeper
	rm -rf $(CURDIR)/services/web-importer/src/lib/sensorbucket
	bunx @hey-api/openapi-ts \
		-i $(CURDIR)/tools/oathkeeper/bundled_openapi.yaml \
		-o $(CURDIR)/services/web-importer/src/lib/sensorbucket \
		-c @hey-api/client-fetch -p @tanstack/svelte-query
