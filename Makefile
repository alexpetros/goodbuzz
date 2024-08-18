# Adapted from https://templ.guide/commands-and-tools/live-reload-with-other-tools/

PORT=3000
BINARY_NAME="buzzer"

.PHONY: live
live:
	@make -j2 live/templ live/air

.PHONY: live/templ
live/templ:
	@templ generate --watch --proxy="http://localhost:$(PORT)" --open-browser=false -v

.PHONY: live/air
live/air:
	@air \
			--tmp_dir "dist" \
			--build.cmd "go build -o dist/tmp/$(BINARY_NAME)" \
			--build.bin "dist/tmp/$(BINARY_NAME)" \
			--build.delay "100" \
			--build.include_ext "go" \
			--build.stop_on_error "false" \
			--misc.clean_on_exit "true"

.PHONY: clean
clean:
	rm -rf ./dist
