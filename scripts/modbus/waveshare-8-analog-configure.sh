#!/usr/bin/env bash

set -euo pipefail

GATEWAY="192.168.2.190"

MODULE_ID=3

echo "Waveshare Modbus RTU Analog Input 8CH"
echo "Gateway: $GATEWAY"
echo

echo "Configuring channels 1-7 as voltage, 8 as current"

mbpoll -m tcp \
    -a "$MODULE_ID" \
    -t 4 \
    -0 \
    -r 4096 \
    -1 \
    "$GATEWAY" \
    0 0 0 0 0 0 0 2

echo
echo "Verifying channel configuration"

mbpoll -m tcp \
    -a "$MODULE_ID" \
    -t 4 \
    -0 \
    -r 4096 \
    -c 8 \
    -1 \
    "$GATEWAY"