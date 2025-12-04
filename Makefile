.PHONY: help serve build clean optimize deploy

help:
	@echo "Available commands:"
	@echo "  make serve     - Start Hugo development server"
	@echo "  make build     - Build production site"
	@echo "  make clean     - Remove public/ directory"
	@echo "  make optimize  - Optimize images"
	@echo "  make deploy    - Build and deploy to GitHub Pages"

serve:
	hugo server -D --disableFastRender

build:
	hugo --minify

clean:
	rm -rf public/

optimize:
	./scripts/image-processing.sh

deploy: clean build
	@echo "Build complete. Deploy to GitHub Pages manually or via CI/CD."
