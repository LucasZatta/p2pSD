#!/bin/bash

# Script para parar os peers do Cenário 1 - Teste B

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Parando peers..."

if [ -f .peer_a.pid ]; then
    PID=$(cat .peer_a.pid)
    if kill -0 $PID 2>/dev/null; then
        echo "Parando Peer A (PID: $PID)..."
        kill $PID
    fi
    rm -f .peer_a.pid
fi

if [ -f .peer_b.pid ]; then
    PID=$(cat .peer_b.pid)
    if kill -0 $PID 2>/dev/null; then
        echo "Parando Peer B (PID: $PID)..."
        kill $PID
    fi
    rm -f .peer_b.pid
fi

echo "Peers parados!"


