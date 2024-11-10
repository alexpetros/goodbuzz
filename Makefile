# Adapted from https://templ.guide/commands-and-tools/live-reload-with-other-tools/

PORT=3000
PROXY_PORT=8080
BINARY_NAME=buzzer
PROJECT_NAME=goodbuzz
DB_NAME=goodbuzz.db

.PHONY: dev
dev:
	make templ
	go build
	GOODBUZZ_PORT=8080 GOODBUZZ_DEV_MODE=true ./$(PROJECT_NAME)

.PHONY: templ
templ:
	templ generate

.PHONY: prod
prod:
	make templ
	go build
	GOODBUZZ_PORT=8080 ./$(PROJECT_NAME)

.PHONY: build
build:
	go install github.com/a-h/templ/cmd/templ@v0.2.747
	make templ
	go build

.PHONY: install
install:
	go install github.com/a-h/templ/cmd/templ@v0.2.747
	make templ
	go install

.PHONY: deploy
deploy:
	ssh goodbuzz 'cd goodbuzz && git pull && make install && systemctl reestart goodbuzz'


.PHONY: clean
clean:
	find . -name '*_templ.go' | xargs rm
	rm -rf ./dist
	go clean

.PHONY: format
format:
	gofmt -s -w .

.PHONY: reset-db
reset-db:
	rm -f $(DB_NAME)
	rm -f $(DB_NAME)-wal
	rm -f $(DB_NAME)-shm
	sqlite3 $(DB_NAME) < db/schema.sql

.PHONY: upcoming
upcoming: reset-db
	sqlite3 $(DB_NAME) < db/october-online.sql


