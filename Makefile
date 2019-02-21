export GO111MODULE=on

SHELL := /bin/bash

TLS_DIR = cmd/tls
TLS_NAME = osctrl-tls
TLS_CODE = ${TLS_DIR:=/*.go}

ADMIN_DIR = cmd/admin
ADMIN_NAME = osctrl-admin
ADMIN_CODE = ${ADMIN_DIR:=/*.go}

CLI_DIR = cmd/cli
CLI_NAME = osctrl-cli
CLI_CODE = ${CLI_DIR:=/*.go}

DEST ?= /opt/osctrl

OUTPUT = build

.PHONY: all build clean

all: build

# Build code according to caller OS and architecture
build:
	make tls
	make admin
	make cli

# Build TLS endpoint
tls:
	go build -o $(OUTPUT)/$(TLS_NAME) $(TLS_CODE)

# Build Admin UI
admin:
	go build -o $(OUTPUT)/$(ADMIN_NAME) $(ADMIN_CODE)

# Build the CLI
cli:
	go build -o $(OUTPUT)/$(CLI_NAME) $(CLI_CODE)

# Delete all compiled binaries
clean:
	rm -rf $(OUTPUT)/$(TLS_NAME)
	rm -rf $(OUTPUT)/$(ADMIN_NAME)
	rm -rf $(OUTPUT)/$(CLI_NAME)

# Install everything
# optional DEST=destination_path
install:
	make clean
	make build
	make install_tls
	make install_admin
	make install_cli

# Install TLS server and restart service
# optional DEST=destination_path
install_tls:
	sudo systemctl stop $(TLS_NAME)
	sudo cp $(TLS_NAME) $(DEST)
	sudo systemctl start $(TLS_NAME)

# Install Admin server and restart service
# optional DEST=destination_path
install_admin:
	sudo systemctl stop $(ADMIN_NAME)
	sudo cp $(ADMIN_NAME) $(DEST)
	sudo systemctl start $(ADMIN_NAME)	

# Install CLI
# optional DEST=destination_path
install_cli:
	sudo cp $(CLI_NAME) $(DEST)

# Auto-format and simplify the code
GOFMT_ARGS = -l -w -s
gofmt-tls:
	gofmt $(GOFMT_ARGS) ./$(TLS_CODE)

gofmt-cli:
	gofmt $(GOFMT_ARGS) ./$(CLI_CODE)

# Run all tests
test:
	# Install dependencies for TLS
	cd $(TLS_DIR) && go test -i . -v
	# Run TLS tests
	cd $(TLS_DIR) && go test . -v
	# Install dependencies for CLI
	cd $(CLI_DIR) && go test -i . -v
	# Run CLI tests
	cd $(CLI_DIR) && go test . -v