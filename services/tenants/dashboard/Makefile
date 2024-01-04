.PHONY: build run watch
build:
	@qtc views/
	@tailwind --config ../../../pkg/layout/tailwind.config.cjs --input ../../../pkg/layout/style.css --output ../../../pkg/layout/static/style.css
	@make -C dependencies build

run: build
	@STATIC_PATH=../../../pkg/layout/static go run .

watch:
	@reflex -r '\.(go|qtpl|js)$$' -R 'static' -R '\.qtpl\.go$$' -s -- make run
