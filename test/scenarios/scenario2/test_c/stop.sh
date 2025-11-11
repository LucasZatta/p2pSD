#!/bin/bash

# Script para parar os peers do Cenário 2 - Teste C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Parando peers..."

for peer in a b c d; do
    PID_FILE=".peer_${peer}.pid"
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if kill -0 $PID 2>/dev/null; then
            echo "Parando Peer ${peer^^} (PID: $PID)..."
            kill $PID
        fi
        rm -f "$PID_FILE"
    fi
done

echo "Peers parados!"


