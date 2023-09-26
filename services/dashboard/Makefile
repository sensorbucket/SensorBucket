.PHONY: build run watch
build:
	@qtc views/
	@tailwind --config tailwind.config.cjs --input style.css --output static/style.css
	@make -C dependencies build

run: build
	@STATIC_PATH=./static go run .

watch:
	@reflex -r '\.(go|qtpl|js)$$' -R 'static' -R '\.qtpl\.go$$' -s -- make run
