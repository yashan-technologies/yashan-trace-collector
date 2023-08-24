# env defines
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
ARCH_AMD=x86_64
ARCH_ARM=aarch64
OS=$(shell if [ $(GOOS)a != ""a ]; then echo $(GOOS); else echo "linux"; fi)
ARCH=$(shell if [ $(GOARCH)a == "arm64"a ]; then echo $(ARCH_ARM); else echo $(ARCH_AMD); fi)
VERSION=$(shell cat ./VERSION)
GO_VERSION=$(shell go env GOVERSION)
GIT_COMMIT_ID=$(shell git rev-parse HEAD)
GIT_DESCRIBE=$(shell git describe --always)

# go command defines
GO_BUILD=go build
GO_MOD_TIDY=$(go mod tidy -compat 1.19)
GO_BUILD_WITH_INFO=$(GO_BUILD) -ldflags "\
	-X 'ytc/defs/compiledef._appVersion=$(VERSION)' \
	-X 'ytc/defs/compiledef._goVersion=$(GO_VERSION)'\
	-X 'ytc/defs/compiledef._gitCommitID=$(GIT_COMMIT_ID)'\
	-X 'ytc/defs/compiledef._gitDescribe=$(GIT_DESCRIBE)'"

# package defines
PKG_PERFIX=yashan-trace-collector
PKG=$(PKG_PERFIX)-$(VERSION)-$(OS)-$(ARCH).tar.gz

BUILD_PATH=./build
PKG_PATH=$(BUILD_PATH)/$(PKG_PERFIX)
BIN_PATH=$(PKG_PATH)/bin
LOG_PATH=$(PKG_PATH)/log
RESULTS_PATH=$(PKG_PATH)/results

# build defines
BIN_YTCD=$(BUILD_PATH)/ytcd
BIN_YTCCTL=$(BUILD_PATH)/ytcctl
BIN_FILES=$(BIN_YTCCTL) $(BIN_YTCD)

DIR_TO_MAKE=$(BIN_PATH) $(LOG_PATH) $(RESULTS_PATH)
FILE_TO_COPY=./config ./scripts ./static

# functions
clean:
	rm -rf $(BUILD_PATH)

define build_ytcd
	$(GO_BUILD_WITH_INFO) -o $(BIN_YTCD) ./cmd/ytcd/*.go
endef

define build_ytcctl
	$(GO_BUILD_WITH_INFO) -o $(BIN_YTCCTL) ./cmd/ytcctl/*.go
endef

go_build: 
	$(GO_MOD_TIDY)
	$(call build_ytcd)
	$(call build_ytcctl)

build: go_build
	@mkdir -p $(DIR_TO_MAKE) 
	@cp -r $(FILE_TO_COPY) $(PKG_PATH)
	@mv $(BIN_FILES) $(BIN_PATH)
	@> $(LOG_PATH)/ytcctl.log
	@cd $(PKG_PATH);ln -s ./bin/ytcctl ./ytcctl
	@cd $(BUILD_PATH);tar -cvzf $(PKG) $(PKG_PERFIX)/

force: clean build