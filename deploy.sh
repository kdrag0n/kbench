#!/usr/bin/env bash

set -eufo pipefail

GOOS=linux GOARCH=arm64 go build -ldflags="-s -w"
adb shell mkdir -p /data/local/tmp/kb
adb push subtests kbench deploy.stage2.sh /data/local/tmp/kb > /dev/null
adb shell chmod 755 /data/local/tmp/kb/deploy.stage2.sh
adb shell su -c /data/local/tmp/kb/deploy.stage2.sh "$@" || true
adb shell rm -f /data/local/tmp/kb/deploy.stage2.sh
