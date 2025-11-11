#!/bin/bash

# Script para executar Cenário 1 - Teste C
# 2 peers, 10MB, bloco 1KB

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
PEER_BIN="$PROJECT_DIR/bin/peer"

cd "$SCRIPT_DIR"

echo "========================================"
echo "  Cenário 1 - Teste C"
echo "  2 peers, 10MB, bloco 1KB"
echo "========================================"
echo ""

# Verifica se o binário existe
if [ ! -f "$PEER_BIN" ]; then
    echo "Erro: binário peer não encontrado em $PEER_BIN"
    echo "Execute: go build -o bin/peer ./cmd/peer"
    exit 1
fi

# Verifica se arquivo existe
if [ ! -f "$SCRIPT_DIR/../../../files/scenario1_file_c.bin" ]; then
    echo "Erro: arquivo de teste não encontrado"
    echo "Execute: cd test && ./genfiles.sh"
    exit 1
fi

# Cria diretórios
mkdir -p logs downloads/peer_a downloads/peer_b

# Limpa logs anteriores
rm -f logs/*.log
rm -rf downloads/peer_b/*

echo "Iniciando Peer A (Seeder) na porta 8001..."
$PEER_BIN -config peers/peer_a.json &
PEER_A_PID=$!
echo "Peer A PID: $PEER_A_PID"

# Aguarda peer A iniciar
sleep 2

echo "Iniciando Peer B (Leecher) na porta 8002..."
$PEER_BIN -config peers/peer_b.json &
PEER_B_PID=$!
echo "Peer B PID: $PEER_B_PID"

# Salva PIDs
echo $PEER_A_PID > .peer_a.pid
echo $PEER_B_PID > .peer_b.pid

echo ""
echo "Peers iniciados!"
echo "- Peer A (Seeder): PID $PEER_A_PID"
echo "- Peer B (Leecher): PID $PEER_B_PID"
echo ""
echo "Logs:"
echo "  tail -f logs/peer_a.log"
echo "  tail -f logs/peer_b.log"
echo ""
echo "Para parar: ./stop.sh"
echo "Para verificar: ./verify.sh"

