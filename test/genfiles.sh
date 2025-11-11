#!/bin/bash

# Script para gerar arquivos de teste para os cenários P2P

set -e  # Para em caso de erro

# Diretórios
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
TEST_DIR="$PROJECT_DIR/test"
FILES_DIR="$TEST_DIR/files"
BIN_DIR="$PROJECT_DIR/bin"

# Cria diretório de arquivos se não existir
mkdir -p "$FILES_DIR"

# Verifica se genfile existe
if [ ! -f "$BIN_DIR/genfile" ]; then
    echo "Erro: genfile não encontrado em $BIN_DIR"
    echo "Execute: go build -o bin/genfile ./cmd/genfile"
    exit 1
fi

echo "========================================"
echo "  Gerando Arquivos de Teste P2P"
echo "========================================"
echo ""

# Cenário 1: 2 peers, bloco 1KB
echo "=== Cenário 1: 2 peers, bloco 1KB ==="

echo "Gerando File A (10KB)..."
$BIN_DIR/genfile -output "$FILES_DIR/scenario1_file_a.bin" -size 10KB -block-size 1024

echo "Gerando File B (1MB)..."
$BIN_DIR/genfile -output "$FILES_DIR/scenario1_file_b.bin" -size 1MB -block-size 1024

echo "Gerando File C (10MB)..."
$BIN_DIR/genfile -output "$FILES_DIR/scenario1_file_c.bin" -size 10MB -block-size 1024

echo ""

# Cenário 2: 4 peers, bloco 4KB
echo "=== Cenário 2: 4 peers, bloco 4KB ==="

echo "Gerando File A (20KB)..."
$BIN_DIR/genfile -output "$FILES_DIR/scenario2_file_a.bin" -size 20KB -block-size 4096

echo "Gerando File B (5MB)..."
$BIN_DIR/genfile -output "$FILES_DIR/scenario2_file_b.bin" -size 5MB -block-size 4096

echo "Gerando File C (20MB)..."
$BIN_DIR/genfile -output "$FILES_DIR/scenario2_file_c.bin" -size 20MB -block-size 4096

echo ""
echo "========================================"
echo "  ✓ Todos os arquivos gerados!"
echo "========================================"
echo ""
echo "Arquivos criados em: $FILES_DIR"
echo ""
echo "Lista de arquivos:"
ls -lh "$FILES_DIR"/*.bin | awk '{print $9, "-", $5}'
echo ""
echo "Metadados:"
ls -lh "$FILES_DIR"/*.meta.json | awk '{print $9}'

