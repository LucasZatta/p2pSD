#!/bin/bash

# Script para verificar resultado do Cenário 2 - Teste A

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "========================================"
echo "  Verificação - Cenário 2 Teste A"
echo "========================================"
echo ""

ORIGINAL_FILE="../../../files/scenario2_file_a.bin"
ORIGINAL_SIZE=$(stat -f%z "$ORIGINAL_FILE" 2>/dev/null || stat -c%s "$ORIGINAL_FILE" 2>/dev/null)
ORIGINAL_HASH=$(sha256sum "$ORIGINAL_FILE" | awk '{print $1}')

ALL_PASSED=true

# Verifica cada leecher (B, C, D)
for peer in b c d; do
    PEER_UPPER=$(echo $peer | tr '[:lower:]' '[:upper:]')
    DOWNLOADED_FILE="downloads/scenario2_file_a.bin"
    
    echo "----------------------------------------"
    echo "Verificando Peer $PEER_UPPER:"
    echo ""
    
    if [ ! -f "$DOWNLOADED_FILE" ]; then
        echo "❌ FALHA: Arquivo não encontrado em $DOWNLOADED_FILE"
        ALL_PASSED=false
        continue
    fi
    
    DOWNLOADED_SIZE=$(stat -f%z "$DOWNLOADED_FILE" 2>/dev/null || stat -c%s "$DOWNLOADED_FILE" 2>/dev/null)
    DOWNLOADED_HASH=$(sha256sum "$DOWNLOADED_FILE" | awk '{print $1}')
    
    echo "Tamanho: $DOWNLOADED_SIZE bytes"
    
    if [ "$ORIGINAL_SIZE" != "$DOWNLOADED_SIZE" ]; then
        echo "❌ FALHA: Tamanho incorreto (esperado: $ORIGINAL_SIZE)"
        ALL_PASSED=false
        continue
    fi
    
    if [ "$ORIGINAL_HASH" != "$DOWNLOADED_HASH" ]; then
        echo "❌ FALHA: Checksum incorreto"
        ALL_PASSED=false
        continue
    fi
    
    echo "✓ Peer $PEER_UPPER: Download válido"
    
    # Mostra estatísticas do log
    if [ -f "logs/peer_${peer}.log" ]; then
        echo ""
        grep "ESTATÍSTICAS\|Arquivo:\|Tamanho:\|Blocos:\|Tempo:\|Throughput:" "logs/peer_${peer}.log" | tail -6
    fi
done

echo ""
echo "========================================"
if [ "$ALL_PASSED" = true ]; then
    echo "  ✓ TESTE PASSOU!"
else
    echo "  ❌ TESTE FALHOU!"
fi
echo "========================================"

