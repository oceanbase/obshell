include Makefile.common

.PHONY: pre-build build bindata obshell rpm buildsucc

default: clean fmt pre-build build

pre-build: bindata swagger

bindata: get
	@if command -v go-bindata > /dev/null 2>&1; then \
		GO_BINDATA=go-bindata; \
	elif [ -f "$(GOPATH_BIN)/go-bindata" ]; then \
		GO_BINDATA="$(GOPATH_BIN)/go-bindata"; \
	else \
		echo "Error: go-bindata not found. Please run 'make get' first."; \
		exit 1; \
	fi; \
	cd ob && $$GO_BINDATA -o agent/bindata/bindata.go -pkg bindata agent/assets/...;cd ..; \
	cd seekdb && $$GO_BINDATA -o agent/bindata/bindata.go -pkg bindata agent/assets/...;cd ..

swagger: get
	@if command -v swag > /dev/null 2>&1; then \
		SWAG=swag; \
	elif [ -f "$(GOPATH_BIN)/swag" ]; then \
		SWAG="$(GOPATH_BIN)/swag"; \
	else \
		echo "Error: swag not found. Please run 'make get' first."; \
		exit 1; \
	fi; \
	cd ob && $$SWAG init -g agent/api/agent_route.go -o agent/api/docs --instanceName ob_swagger;cd ..; \
	cd seekdb && $$SWAG init -g agent/api/agent_route.go -o agent/api/docs --instanceName seekdb_swagger;cd ..

build: build-debug

build-debug: set-debug-flags obshell buildsucc

build-release: set-release-flags obshell buildsucc

build-with-swagger: enable-swagger build-release

build-for-test: pre-build enable-swagger set-disable-encryption-flags build-debug

frontend-dep:
	npm i -g pnpm@8

seekdb-frontend-build:
	cd seekdb/frontend && pnpm i && pnpm build && cd ../

seekdb-frontend-build-tester:
	cd seekdb/frontend && pnpm i && pnpm build:tester && cd ../

ob-frontend-build:
	cd ob/frontend && pnpm i && pnpm build && cd ../

ob-frontend-build-tester:
	cd ob/frontend && pnpm i && pnpm build:tester && cd ../

rpm:
	@if [ "$(UNAME_S)" = "Darwin" ]; then \
		echo "Error: RPM build is not supported on macOS. Please use a Linux system or Docker to build RPM packages."; \
		exit 1; \
	fi
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
	$(GO) install github.com/swaggo/swag/cmd/swag@v1.16.4


vet:
	go vet $$(go list ./...)

clean:
	rm -rf $(GOCOVERAGE_FILE)
	rm -rf tests/mock/*
	rm -rf bin/ob_mgragent bin/ob_monagent bin/ob_agentctl bin/ob_agentd
	$(GO) clean -i ./...

init: pre-build pre-test tidy
