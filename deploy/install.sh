#!/usr/bin/env sh
set -e

active=$(systemctl is-active sinkhole.service)
if [ "$active" = "active" ]; then
  systemctl stop sinkhole.service
fi

mkdir -p sink/bin
mv hosts sink/
mv hole sink/bin/

enabled=$(systemctl is-enabled sinkhole.service)
if [ "$enabled" = "enabled" ]; then
  systemctl disable sinkhole.service
fi

mv sinkhole.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable sinkhole.service

echo "run 'sudo systemctl start sinkhole.service' to start sinkhole"
