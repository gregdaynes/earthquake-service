.PHONY: run
run:
	go run ./cmd/web

.PHONY: dev
dev:
	go run ./cmd/web -debug -host=0.0.0.0 -port=4000

.PHONY: tidy
tidy:
	@echo 'Formatting .go files...'
	go fmt ./...
	@echo 'Tidying module dependencies...'
	go mod tidy
	@echo 'Verifying and vendoring module dependencies...'
	go mod verify
	go mod vendor

.PHONY: audit
audit:
	@echo 'Checking module dependencies'
	go mod tidy -diff
	go mod verify
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

.PHONY: build
build:
	go build -a -o ./dist/quakes.local ./cmd/web 

.PHONY: build-linux-amd64
build-linux-amd64:
	GOARCH=amd64 GOOS=linux go build -a -o ./dist/quakes.linux_amd64 ./cmd/web/

