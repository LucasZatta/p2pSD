#!/bin/bash

# Script para executar Cenário 2 - Teste B
# 4 peers, 5MB, bloco 4KB

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
PEER_BIN="$PROJECT_DIR/bin/peer"

cd "$SCRIPT_DIR"

echo "========================================"
echo "  Cenário 2 - Teste B"
echo "  4 peers, 5MB, bloco 4KB"
echo "========================================"
echo ""

# Verifica se o binário existe
if [ ! -f "$PEER_BIN" ]; then
    echo "Erro: binário peer não encontrado em $PEER_BIN"
    echo "Execute: go build -o bin/peer ./cmd/peer"
    exit 1
fi

# Verifica se arquivo existe
if [ ! -f "$SCRIPT_DIR/../../../files/scenario2_file_b.bin" ]; then
    echo "Erro: arquivo de teste não encontrado"
    echo "Execute: cd test && ./genfiles.sh"
    exit 1
fi

# Cria diretórios
mkdir -p logs downloads/peer_a downloads/peer_b downloads/peer_c downloads/peer_d

# Limpa logs e downloads anteriores
rm -f logs/*.log
rm -rf downloads/peer_b/* downloads/peer_c/* downloads/peer_d/*

echo "Iniciando Peer A (Seeder) na porta 8101..."
$PEER_BIN -config peers/peer_a.json &
PEER_A_PID=$!
echo "Peer A PID: $PEER_A_PID"

# Aguarda peer A iniciar
sleep 2

echo "Iniciando Peer B (Leecher) na porta 8102..."
$PEER_BIN -config peers/peer_b.json &
PEER_B_PID=$!
echo "Peer B PID: $PEER_B_PID"

sleep 1

echo "Iniciando Peer C (Leecher) na porta 8103..."
$PEER_BIN -config peers/peer_c.json &
PEER_C_PID=$!
echo "Peer C PID: $PEER_C_PID"

sleep 1

echo "Iniciando Peer D (Leecher) na porta 8104..."
$PEER_BIN -config peers/peer_d.json &
PEER_D_PID=$!
echo "Peer D PID: $PEER_D_PID"

# Salva PIDs
echo $PEER_A_PID > .peer_a.pid
echo $PEER_B_PID > .peer_b.pid
echo $PEER_C_PID > .peer_c.pid
echo $PEER_D_PID > .peer_d.pid

echo ""
echo "Peers iniciados!"
echo "- Peer A (Seeder): PID $PEER_A_PID"
echo "- Peer B (Leecher): PID $PEER_B_PID"
echo "- Peer C (Leecher): PID $PEER_C_PID"
echo "- Peer D (Leecher): PID $PEER_D_PID"
echo ""
echo "Logs:"
echo "  tail -f logs/peer_a.log"
echo "  tail -f logs/peer_b.log"
echo "  tail -f logs/peer_c.log"
echo "  tail -f logs/peer_d.log"
echo ""
echo "Para parar: ./stop.sh"
echo "Para verificar: ./verify.sh"

