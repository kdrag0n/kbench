#!/system/bin/sh

set -eufo pipefail

cd /data/local/tmp/kb
./kbench "$@"
