GIT_COMMIT 	= $(shell git rev-parse HEAD)
GIT_SHA    	= $(shell git rev-parse --short HEAD)
GIT_TAG    	= $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY  	= $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

GOPATH		= $(shell go env GOPATH)
GOLINTCI	= $(GOPATH)/bin/golangci-lint 

ifndef VERSION
	VERSION = $(GIT_TAG)
endif

dirty = "dirty"

lint:
	@echo "running golangci-lint..."
	@$(GOLINTCI) run ./...



release: check lint
	@echo "pushing to origin/develop"
	$(shell git push origin develop)
	@echo "pushing to origin/${GIT_TAG}'"
	$(shell git push origin ${GIT_TAG})

check:
	@echo "checking status..."
ifeq ($(GIT_DIRTY), dirty)
	$(error git state is not clean)
endif


head:
	@echo "Git short head:	   $(GIT_SHA)"

info:
	@echo "Version:           ${VERSION}"
	@echo "Git Tag:           ${GIT_TAG}"
	@echo "Git Commit:        ${GIT_COMMIT}"
	@echo "Git Tree State:    ${GIT_DIRTY}"
