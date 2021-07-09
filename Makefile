HOSTNAME=registry.terraform.io
NAMESPACE=curve-technology
NAME=confluent-schema-registry
BINARY=terraform-provider-${NAME}
VERSION=0.1.0
OS_ARCH=darwin_amd64

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)


.PHONY: help
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${YELLOW}%-18s${GREEN}%s${RESET}\n", $$1, $$2}' $(MAKEFILE_LIST)


.PHONY: build
build: ## Build project and put output binary in /bin folder
	@go build -o bin/${BINARY}_${VERSION}_darwin_amd64 main.go


.PHONY: release
release: ## Build project for multiple target platforms
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64 main.go
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386 main.go
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64 main.go


# .PHONY: install
# install: build ## Build project and install it in terraform plugin folder (MacOS only)
# 	@mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
# 	@cp bin/${BINARY}_${VERSION}_darwin_amd64 ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}


.PHONY: test
test: ## Run the tests of the project
	@go test -race ./...


.PHONY: coverage
coverage: ## Run the tests of the project and print out coverage
	@go test -cover ./...


.PHONY: coverage-report
coverage-report: ## Run the tests of the project and show line by line coverage in the browser
	@go test -coverprofile=coverage.txt ./...
	@go tool cover -html=coverage.txt


.PHONY: lint
lint: ## Run linters
	@gofmt -l .
	@go vet ./...


.PHONY: clean
clean: ## Remove temporary and build related files
	@rm -f coverage.txt
	@rm -f bin/*
