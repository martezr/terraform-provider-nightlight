HOSTNAME   = registry.terraform.io
NAMESPACE  = martezr
NAME       = nightlight
VERSION    ?= 0.1.0
OS_ARCH    ?= $(shell go env GOOS)_$(shell go env GOARCH)
PLUGIN_DIR = ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)

default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

install-local: build
	mkdir -p $(PLUGIN_DIR)
	cp $(shell go env GOPATH)/bin/terraform-provider-$(NAME) $(PLUGIN_DIR)/terraform-provider-$(NAME)_v$(VERSION)
	@echo "Provider installed to $(PLUGIN_DIR)"
	@echo "Add the following to your Terraform config to use it:"
	@echo ""
	@echo "  terraform {"
	@echo "    required_providers {"
	@echo "      $(NAME) = {"
	@echo "        source  = \"$(HOSTNAME)/$(NAMESPACE)/$(NAME)\""
	@echo "        version = \"$(VERSION)\""
	@echo "      }"
	@echo "    }"
	@echo "  }"

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build install install-local generate
