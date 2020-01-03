GIT_DIRTY  	= $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

GOPATH		= $(shell go env GOPATH)
GOLINTCI	= $(GOPATH)/bin/golangci-lint 
MISSPELL	= $(GOPATH)/bin/misspell

GO 		 = go
TIMEOUT  = 20
PKGS     = $(or $(PKG),$(shell env GO111MODULE=on $(GO) list ./...))
TESTPKGS = $(shell env GO111MODULE=on $(GO) list -f \
            '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' \
            $(PKGS))

Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

DESCRIBE           := $(shell git describe --match "v*" --always --tags)
DESCRIBE_PARTS     := $(subst -, ,$(DESCRIBE))

VERSION_TAG        := $(word 1,$(DESCRIBE_PARTS))
COMMITS_SINCE_TAG  := $(word 2,$(DESCRIBE_PARTS))

VERSION            := $(subst v,,$(VERSION_TAG))
VERSION_PARTS      := $(subst ., ,$(VERSION))

MAJOR              := $(word 1,$(VERSION_PARTS))
MINOR              := $(word 2,$(VERSION_PARTS))
MICRO              := $(word 3,$(VERSION_PARTS))

NEXT_MAJOR         := $(shell echo $$(($(MAJOR)+1)))
NEXT_MINOR         := $(shell echo $$(($(MINOR)+1)))
NEXT_MICRO          = $(shell echo $$(($(MICRO)+$(COMMITS_SINCE_TAG))))

ifeq ($(strip $(COMMITS_SINCE_TAG)),)
CURRENT_VERSION_MICRO := $(MAJOR).$(MINOR).$(MICRO)
CURRENT_VERSION_MINOR := $(CURRENT_VERSION_MICRO)
CURRENT_VERSION_MAJOR := $(CURRENT_VERSION_MICRO)
else
CURRENT_VERSION_MICRO := $(MAJOR).$(MINOR).$(NEXT_MICRO)
CURRENT_VERSION_MINOR := $(MAJOR).$(NEXT_MINOR).0
CURRENT_VERSION_MAJOR := $(NEXT_MAJOR).0.0
endif

DATE                = $(shell date +'%d.%m.%Y')
TIME                = $(shell date +'%H:%M:%S')
COMMIT             := $(shell git rev-parse HEAD)
AUTHOR             := $(firstword $(subst @, ,$(shell git show --format="%aE" $(COMMIT))))
BRANCH_NAME        := $(shell git rev-parse --abbrev-ref HEAD)

TAG_MESSAGE         = "$(TIME) $(DATE) $(AUTHOR) $(BRANCH_NAME)"
COMMIT_MESSAGE     := $(shell git log --format=%B -n 1 $(COMMIT))

CURRENT_TAG_MICRO  := "v$(CURRENT_VERSION_MICRO)"
CURRENT_TAG_MINOR  := "v$(CURRENT_VERSION_MINOR)"
CURRENT_TAG_MAJOR  := "v$(CURRENT_VERSION_MAJOR)"

dirty = "dirty"

RELEASE_TARGETS = release-patch release-minor release-major
.PHONY: $(RELEASE_TARGETS) release
release-patch: CURRENT_TAG=$(CURRENT_TAG_MICRO)
release-minor: CURRENT_TAG=$(CURRENT_TAG_MINOR)
release-major: CURRENT_TAG=$(CURRENT_TAG_MAJOR)
$(RELEASE_TARGETS): release

## release a version
release: test-race lint commit
	$(info $(M) creating tag: '${CURRENT_TAG}'…)
	$(shell git tag -a ${CURRENT_TAG} -m ${TAG_MESSAGE})
	$(info $(M) pushing to origin/develop…)
	$(shell git push origin develop)
	$(info $(M) pushing to origin/${CURRENT_TAG}…)
	$(shell git push origin ${CURRENT_TAG})


## check git status
.PHONY: check
check:
	$(info $(M) checking git status…)
ifeq ($(GIT_DIRTY), dirty)
	$(error git state is not clean)
endif

.PHONY: commit
commit:
ifeq ($(GIT_DIRTY), dirty)
	$(info $(M) preparing commit…) @
	## $(shell git add --all)
	## $(info $(M) added all files...)
	$(shell git commit -am "$(COMMIT_MESSAGE)")
endif

.PHONY: info
info:
	@echo "Git Commit:        ${COMMIT}"
	@echo "Git Tree State:    ${GIT_DIRTY}"

## Todos
.PHONY: todo
todo:
	@grep \
		--exclude-dir=vendor \
		--exclude=Makefile \
		--exclude=*.swp \
		--text \
		--color \
		-nRo -E ' TODO:.*|SkipNow' .

## Tests
TEST_TARGETS := test-default test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) test
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
test:
	$(info $(M) running $(NAME:%=% )tests…) @ ## Run tests
	$Q $(GO) test -timeout $(TIMEOUT)s $(ARGS) $(TESTPKGS)

## Format
.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	$Q $(GO) fmt $(PKGS)

## Linters
.PHONY: lint
lint:
	$(info $(M) running linters…)
	@$(GOLINTCI) run ./...
	@$(MISSPELL) -error **/*
