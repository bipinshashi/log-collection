.PHONY: test run build

test:
	go test -v ./... -coverprofile=coverage.out

run:
	docker compose up -d

build:
	docker compose build