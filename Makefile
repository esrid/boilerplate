
esbuild:
	esbuild --bundle --minify --outdir=./web/static/js/ --watch ./web/sources/*.ts

tailwind:
	tailwindcss -i ./web/sources/app.css -o ./web/static/css/style.css --watch --optimize

dev-js:
	make esbuild

dev-css:
	make tailwind

dev: 
	@echo "Run 'make dev-js' and 'make dev-css' in separate terminals for development"

build:
	@echo "Building project..."
	esbuild --bundle --minify --outdir=./web/static/js/ ./web/sources/*.ts
	tailwindcss -i ./web/sources/app.css -o ./web/static/css/style.css --optimize

clean:
	@echo "Cleaning build artifacts..."
	rm -rf ./web/static/js/*
	rm -rf ./web/static/css/*

docker-build:
	docker build -t breakit .

docker-run:
	docker run -p 8080:8080 breakit

.PHONY: esbuild tailwind dev-js dev-css dev build clean docker-build docker-run

rename: 
	 find . -type f -name '*.go.tpl' -exec sed -i '' 's/yourapp/{{projectName}}/g' {} +
