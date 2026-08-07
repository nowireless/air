#! /usr/bin/env bash

# Dew point sensor
# - Address: 1


mbpoll -m tcp \
    -m tcp \
    -a 1 \
    -r 1 \
    -c 5 \
    -t 4 \
    192.168.9.8