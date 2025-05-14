#!/bin/bash
set -e

VERSION=$(curl -s https://api.github.com/repos/fastly/cli/releases/latest | jq -r .tag_name)
wget https://github.com/fastly/cli/releases/download/${VERSION}/fastly_${VERSION}_linux-amd64.tar.gz
tar -xzf fastly_${VERSION}_linux-amd64.tar.gz
mv fastly /usr/local/bin/

# Clean up archive and extracted directory if needed
rm -f fastly_${VERSION}_linux-amd64.tar.gz
rm -rf fastly_${VERSION}_linux-amd64

# Verify installation
fastly version

# Run project setup/build commands
go mod tidy
fastly compute build
fastly compute pack --wasm-binary=bin/main.wasm
