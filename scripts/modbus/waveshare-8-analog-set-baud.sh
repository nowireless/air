#!/usr/bin/env bash

set -euo pipefail

GATEWAY="192.168.2.190"

OLD_ID=1
NEW_ID=3

OLD_BAUD=9600
NEW_BAUD=19200

# Waveshare register addresses (zero-based)
REG_DEVICE_ID=16384  # 0x4000
REG_UART=8192        # 0x2000

# UART configuration:
# 0x0001 = 9600, no parity
# 0x0002 = 19200, no parity
UART_CONFIG=2

read_register() {
    local id="$1"
    local register="$2"

    mbpoll -m tcp \
        -a "$id" \
        -t 4 \
        -0 \
        -r "$register" \
        -c 1 \
        -1 \
        "$GATEWAY"
}

write_register() {
    local id="$1"
    local register="$2"
    local value="$3"

    mbpoll -m tcp \
        -a "$id" \
        -t 4 \
        -0 \
        -r "$register" \
        -1 \
        "$GATEWAY" "$value"
}

echo "Waveshare Modbus RTU Analog Input 8CH"
echo "Gateway: $GATEWAY"
echo

read -rp "Configure gateway for $OLD_BAUD baud, then press Enter..."

echo
echo "Checking communication with slave ID $OLD_ID..."

read_register "$OLD_ID" "$REG_DEVICE_ID"

echo
echo "Changing slave ID: $OLD_ID -> $NEW_ID"

write_register "$OLD_ID" "$REG_DEVICE_ID" "$NEW_ID"

echo
echo "Verifying new slave ID..."

read_register "$NEW_ID" "$REG_DEVICE_ID"

echo
echo "Changing baud rate: $OLD_BAUD -> $NEW_BAUD"

# IMPORTANT: Use the NEW slave ID here!
write_register "$NEW_ID" "$REG_UART" "$UART_CONFIG" || {
    echo "Warning: Baud rate write timed out."
    echo "The device may already have switched to $NEW_BAUD."
}

echo

read -rp "Configure gateway for $NEW_BAUD baud, then press Enter..."

echo
echo "Verifying UART configuration..."

read_register "$NEW_ID" "$REG_UART"

echo
echo "Reading all 8 analog inputs..."

mbpoll -m tcp \
    -a "$NEW_ID" \
    -t 3 \
    -0 \
    -r 0 \
    -c 8 \
    -1 \
    "$GATEWAY"

echo
echo "Configuration complete!"