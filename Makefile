.PHONY: help serve build build-debug build-hugo-only clean optimize deploy check-seo check-redirect

help:
	@echo "Available commands:"
	@echo "  make serve          - Start Hugo development server"
	@echo "  make build         - Build production site (with image optimization)"
	@echo "  make build-hugo-only - Build without image preprocessing (if ImageMagick unavailable)"
	@echo "  make build-debug   - Build with template metrics"
	@echo "  make clean         - Remove public/ directory"
	@echo "  make optimize      - Optimize images"
	@echo "  make check-seo     - Validate SEO in the generated public/ site"
	@echo "  make check-redirect - Test the Cloudflare www redirect"
	@echo "  make deploy        - Build and deploy to GitHub Pages"

serve:
	hugo server -D --disableFastRender

build: optimize
	hugo --minify --gc --cleanDestinationDir
	$(MAKE) check-seo

build-debug: optimize
	hugo --minify --gc --cleanDestinationDir --templateMetrics --templateMetricsHints
	$(MAKE) check-seo

build-hugo-only:
	hugo --minify --gc --cleanDestinationDir
	$(MAKE) check-seo

check-seo:
	python3 scripts/check-seo.py public

check-redirect:
	node --test cloudflare/www-redirect/redirect.test.mjs

clean:
	rm -rf public/

optimize:
	./scripts/image-processing.sh

deploy: clean build
	@echo "Build complete. Deploy to Cloudflare Pages manually or via CI/CD."
