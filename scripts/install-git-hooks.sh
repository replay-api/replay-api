#!/bin/bash
#
# Script para instalar git hooks
#

set -e

# Cores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "${GREEN}Installing git hooks...${NC}"

# Verifica se estamos em um repositório git
if [ ! -d .git ]; then
    echo "${YELLOW}Error: Not a git repository${NC}"
    exit 1
fi

# Cria diretório de hooks se não existir
mkdir -p .git/hooks

# Copia hooks do diretório .githooks
if [ -d .githooks ]; then
    for hook in .githooks/*; do
        if [ -f "$hook" ]; then
            hook_name=$(basename "$hook")
            cp "$hook" ".git/hooks/$hook_name"
            chmod +x ".git/hooks/$hook_name"
            echo "${GREEN}✓ Installed $hook_name${NC}"
        fi
    done
else
    echo "${YELLOW}Warning: .githooks directory not found${NC}"
fi

echo "${GREEN}Git hooks installed successfully!${NC}"
echo "${YELLOW}Note: Make sure you have the following tools installed:${NC}"
echo "  - gofmt (comes with Go)"
echo "  - goimports (go install golang.org/x/tools/cmd/goimports@latest)"
echo "  - golangci-lint (optional, but recommended)"

