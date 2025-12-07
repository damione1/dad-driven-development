.PHONY: help serve build build-debug build-hugo-only clean optimize deploy

help:
	@echo "Available commands:"
	@echo "  make serve          - Start Hugo development server"
	@echo "  make build         - Build production site (with image optimization)"
	@echo "  make build-hugo-only - Build without image preprocessing (if ImageMagick unavailable)"
	@echo "  make build-debug   - Build with template metrics"
	@echo "  make clean         - Remove public/ directory"
	@echo "  make optimize      - Optimize images"
	@echo "  make deploy        - Build and deploy to GitHub Pages"

serve:
	hugo server -D --disableFastRender

build: optimize
	hugo --minify --gc --cleanDestinationDir

build-debug: optimize
	hugo --minify --gc --cleanDestinationDir --templateMetrics --templateMetricsHints

build-hugo-only:
	hugo --minify --gc --cleanDestinationDir

clean:
	rm -rf public/

optimize:
	./scripts/image-processing.sh

deploy: clean build
	@echo "Build complete. Deploy to Cloudflare Pages manually or via CI/CD."
