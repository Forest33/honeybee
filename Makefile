#!make
.SILENT:

run: app

app:
	docker compose build
	docker compose up -d --force-recreate


test:
	go clean --testcache
	go test ./...

deps:
	go mod download && go mod tidy
