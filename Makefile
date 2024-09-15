# Adapted from https://templ.guide/commands-and-tools/live-reload-with-other-tools/

PORT=3000
PROXY_PORT=8080
BINARY_NAME="buzzer"
PROJECT_NAME="goodbuzz"

.PHONY: live
live:
	BUZZER_PROXY_PORT=$(PROXY_PORT) make -j2 live/templ live/air
	@echo "Development mode is live\n Access application at http://localhost:$(PROXY_PORT)"

.PHONY: templ
templ:
	templ generate

.PHONY: prod
prod:
	make templ
	go build
	BUZZER_PORT=8080 ./$(PROJECT_NAME)

.PHONY: live/templ
live/templ:
	@templ generate --watch \
		--proxy="http://localhost:$(PORT)" \
		--proxyport="$(PROXY_PORT)" \
		--open-browser=false -v

.PHONY: live/air
live/air:
	@air \
			--tmp_dir "dist" \
			--build.cmd "go build -o dist/tmp/$(BINARY_NAME)" \
			--build.bin "dist/tmp/$(BINARY_NAME)" \
			--build.delay "100" \
			--build.include_ext "go,css" \
			--build.stop_on_error "false" \
			--misc.clean_on_exit "true"

.PHONY: clean
clean:
	find . -name '*_templ.go' | xargs rm
	rm -rf ./dist
	go clean

.PHONY: format
format:
	gofmt -s -w .
