#!/bin/bash

# Script para verificar resultado do Cenário 1 - Teste B

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "========================================"
echo "  Verificação - Cenário 1 Teste B"
echo "========================================"
echo ""

ORIGINAL_FILE="../../../files/scenario1_file_b.bin"
DOWNLOADED_FILE="downloads/scenario1_file_b.bin"

# Verifica se arquivo baixado existe
if [ ! -f "$DOWNLOADED_FILE" ]; then
    echo "❌ FALHA: Arquivo baixado não encontrado em $DOWNLOADED_FILE"
    exit 1
fi

echo "Arquivo baixado encontrado: $DOWNLOADED_FILE"
echo ""

# Compara tamanhos
ORIGINAL_SIZE=$(stat -f%z "$ORIGINAL_FILE" 2>/dev/null || stat -c%s "$ORIGINAL_FILE" 2>/dev/null)
DOWNLOADED_SIZE=$(stat -f%z "$DOWNLOADED_FILE" 2>/dev/null || stat -c%s "$DOWNLOADED_FILE" 2>/dev/null)

echo "Tamanho do arquivo original: $ORIGINAL_SIZE bytes"
echo "Tamanho do arquivo baixado:  $DOWNLOADED_SIZE bytes"

if [ "$ORIGINAL_SIZE" != "$DOWNLOADED_SIZE" ]; then
    echo "❌ FALHA: Tamanhos diferentes!"
    exit 1
fi

echo "✓ Tamanhos correspondem"
echo ""

# Compara checksums SHA-256
echo "Calculando checksums SHA-256..."
ORIGINAL_HASH=$(sha256sum "$ORIGINAL_FILE" | awk '{print $1}')
DOWNLOADED_HASH=$(sha256sum "$DOWNLOADED_FILE" | awk '{print $1}')

echo "Hash do arquivo original: $ORIGINAL_HASH"
echo "Hash do arquivo baixado:  $DOWNLOADED_HASH"
echo ""

if [ "$ORIGINAL_HASH" != "$DOWNLOADED_HASH" ]; then
    echo "❌ FALHA: Checksums diferentes!"
    exit 1
fi

echo "✓ Checksums correspondem"
echo ""
echo "========================================"
echo "  ✓ TESTE PASSOU!"
echo "========================================"
echo ""

# Mostra estatísticas dos logs
if [ -f logs/peer_b.log ]; then
    echo "Estatísticas do Peer B:"
    grep "ESTATÍSTICAS\|Arquivo:\|Tamanho:\|Blocos:\|Tempo:\|Throughput:\|Checksum:" logs/peer_b.log | tail -7
fi

