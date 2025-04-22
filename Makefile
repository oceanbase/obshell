include Makefile.common

.PHONY: pre-build build bindata obshell rpm buildsucc

default: clean fmt pre-build build

pre-build: bindata swagger

bindata: get
	go-bindata -o agent/bindata/bindata.go -pkg bindata agent/assets/...

swagger: get
	swag init -g agent/api/agent_route.go -o agent/api/docs

build: build-debug

build-debug: set-debug-flags obshell buildsucc

build-release: set-release-flags obshell buildsucc

build-with-swagger: enable-swagger build-release

build-for-test: pre-build enable-swagger set-disable-encryption-flags build-debug

frontend-dep:
	npm i -g pnpm@7

frontend-build:
	cd frontend && pnpm i && pnpm build && cd ../

rpm:
	cd ./rpm && VERSION=$(VERSION) RELEASE=$(RELEASE) NAME=$(NAME) OBSHELL_RELEASE=$(OBSHELL_RELEASE) rpmbuild -bb obshell.spec

set-disable-encryption-flags:
	@echo Build with encryption disabled flags
	$(eval LDFLAGS += $(LDFLAGS_DISABLE_ENCRYPTION))

set-debug-flags:
	@echo Build with debug flags
	$(eval LDFLAGS += $(LDFLAGS_DEBUG))
	$(eval BUILD_FLAG += $(GO_RACE_FLAG))	

set-release-flags:
	@echo Build with release flags
	$(eval LDFLAGS += $(LDFLAGS_RELEASE))

enable-swagger:
	@echo Build with swagger flags
	$(eval BUILD_FLAG +=-tags swagger)

enable-sm4:
	@echo Build with sm4 flags
	$(eval BUILD_FLAG +=-tags sm4)

obshell:
	$(GO) build $(BUILD_FLAG) -ldflags '$(OBSHELL_LDFLAGS)' -o bin/obshell cmd/main.go

buildsucc:
	@echo Build obshell successfully!

fmt:
	@gofmt -s -w $(filter-out , $(GOFILES))

fmt-check:
	@if [ -z "$(UNFMT_FILES)" ]; then \
		echo "gofmt check passed"; \
		exit 0; \
    else \
    	echo "gofmt check failed, not formatted files:"; \
    	echo "$(UNFMT_FILES)" | tr -s " " "\n"; \
    	exit 1; \
    fi

tidy:
	$(GO) mod tidy

get:
	$(GO) install github.com/go-bindata/go-bindata/...@v3.1.2+incompatible
	$(GO) install github.com/golang/mock/mockgen@v1.6.0
	$(GO) install github.com/swaggo/swag/cmd/swag@latest


vet:
	go vet $$(go list ./...)

clean:
	rm -rf $(GOCOVERAGE_FILE)
	rm -rf tests/mock/*
	rm -rf bin/ob_mgragent bin/ob_monagent bin/ob_agentctl bin/ob_agentd
	$(GO) clean -i ./...

init: pre-build pre-test tidy
