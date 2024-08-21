# Adapted from https://templ.guide/commands-and-tools/live-reload-with-other-tools/

PORT=3000
PROXY_PORT=8080
BINARY_NAME="buzzer"

.PHONY: live
live:
	BUZZER_PROXY_PORT=$(PROXY_PORT) make -j2 live/templ live/air
	@echo "Development mode is live\n Access application at http://localhost:$(PROXY_PORT)"

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

# TODO update this to remove the templ files
.PHONY: clean
clean:
	rm -rf ./dist

.PHONY: format
format:
	gofmt -s -w .
