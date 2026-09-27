#!/usr/bin/env bash

set -euo pipefail

GATEWAY="192.168.2.190"

MODULE_ID=3


echo "Waveshare Modbus RTU Analog Input 8CH"
echo "Gateway: $GATEWAY"
echo

echo
echo "Reading all 8 analog inputs..."

mbpoll -m tcp \
    -a "$MODULE_ID" \
    -t 3 \
    -0 \
    -r 0 \
    -c 8 \
    -1 \
    "$GATEWAY"
