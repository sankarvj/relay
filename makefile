SHELL := /bin/bash

export PROJECT = relay-project

all: relay-api metrics

run: 
	go run ./cmd/relay-api/main.go

worker: 
	go run ./cmd/relay-worker/main.go

admin:
	go run ./cmd/relay-admin/main.go --db-disable-tls=1 useradd admin@example.com gophers

brain: 
	go run ./cmd/relay-slack-brain/main.go

keys:
	go run ./cmd/relay-admin/main.go keygen private.pem

migrate:
	go run ./cmd/relay-admin/main.go --db-disable-tls=1 migrate

seed: migrate
	go run ./cmd/relay-admin/main.go --db-disable-tls=1 seed

crm:
	go run ./cmd/relay-admin/main.go --db-disable-tls=1 crmadd

csm:
	go run ./cmd/relay-admin/main.go --db-disable-tls=1 csmadd

pm:
	go run ./cmd/relay-admin/main.go --db-disable-tls=1 pmadd

em:
	go run ./cmd/relay-admin/main.go --db-disable-tls=1 emadd

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

build:
	env GOOS=linux GOARCH=arm go build ./cmd/relay-api -o bin/application application.go

deploy-api:
	./deploy-api.sh

deploy-worker:
	./deploy-worker.sh
	
deploy-admin:
	./deploy-admin.sh

ngrok:
	../ngrok http 3000

relayrok:
	../ngrok http -hostname=relay.ngrok.io 3000

brainrok:
	../ngrok http 3002
