SHELL := /bin/bash

export PROJECT = relay-project

all: relay-api metrics

run: 
	go run ./cmd/relay-api/main.go

keys:
	go run ./cmd/relay-admin/main.go keygen private.pem

admin:
	go run ./cmd/relay-admin/main.go --db-disable-tls=1 useradd admin@example.com gophers

migrate:
	go run ./cmd/relay-admin/main.go --db-disable-tls=1 migrate

seed: migrate
	go run ./cmd/relay-admin/main.go --db-disable-tls=1 seed

relay-api:
	docker build \
		-f dockerfile.relay-api \
		-t gcr.io/$(PROJECT)/relay-api-amd64:1.0 \
		--build-arg PACKAGE_NAME=relay-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

metrics:
	docker build \
		-f dockerfile.metrics \
		-t gcr.io/$(PROJECT)/metrics-amd64:1.0 \
		--build-arg PACKAGE_NAME=metrics \
		--build-arg PACKAGE_PREFIX=sidecar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

up:
	docker-compose up

down:
	docker-compose down

test:
	go test -mod=vendor ./... -count=1

clean-test:
	go clean -testcache

clean:
	docker system prune -f

stop-all:
	docker stop $(docker ps -aq)

remove-all:
	docker rm $(docker ps -aq)

deps-reset:
	git checkout -- go.mod
	go mod tidy

deps-upgrade:
	go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)

deps-cleancache:
	go clean -modcache
