.PHONY: build run watch
build:
	@qtc transports/webui/views/
	@tailwind --config ./tailwind.config.cjs --input transports/webui/style.css --output transports/webui/static/style.css

run: build
	@go run .

watch:
	@reflex -r '\.(go|qtpl|js)$$' -R 'static' -R '\.qtpl\.go$$' -s -- make run
