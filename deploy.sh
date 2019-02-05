#!/usr/bin/env bash
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" && adb push kbench deploy.stage2.sh /data/local/tmp > /dev/null && adb shell chmod 755 /data/local/tmp/deploy.stage2.sh && adb shell su -c /data/local/tmp/deploy.stage2.sh "$@"
adb shell rm -f /data/local/tmp/deploy.stage2.sh

