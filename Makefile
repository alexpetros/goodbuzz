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
	ssh goodbuzz 'cd goodbuzz && git pull && make install && systemctl restart goodbuzz'

.PHONY: download-db
download-db:
	rm -f goodbuzz.db*
	ssh goodbuzz 'sqlite3 goodbuzz/goodbuzz.db .dump' | sqlite3 goodbuzz.db


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

.PHONY: show-admin
show-admin:
	sqlite3 $(DB_NAME) "select value from settings where key = 'admin_password'"

.PHONY: show-mod
show-mod:
	sqlite3 $(DB_NAME) "select value from settings where key = 'mod_password'"

