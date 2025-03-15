.PHONY: create-bin-dir
create-bin-dir:
	mkdir -p ./bin/$(NAME)

.PHONY: compress
compress:
	tar -czvf ./bin/$(NAME).tar.gz -C ./bin $(NAME)

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -trimpath -o ./bin/$(NAME)/mu main.go

.PHONY: build-linux-x86_64
build-linux-x86_64: GOOS=linux
build-linux-x86_64: GOARCH=amd64
build-linux-x86_64: NAME=mu_Linux_x86_64
build-linux-x86_64: VERSION=dev
build-linux-x86_64: COMMIT=none
build-linux-x86_64: DATE=unknown
build-linux-x86_64: create-bin-dir build compress

.PHONY: build-linux-arm64
build-linux-arm64: GOOS=linux
build-linux-arm64: GOARCH=arm64
build-linux-arm64: NAME=mu_Linux_arm64
build-linux-arm64: VERSION=dev
build-linux-arm64: COMMIT=none
build-linux-arm64: DATE=unknown
build-linux-arm64: create-bin-dir build compress

.PHONY: test
test:
	go test -race ./...

.PHONY: lint
lint:
	golangci-lint run
