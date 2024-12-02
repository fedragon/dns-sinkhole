#!/usr/bin/env sh

# stop sinkhole service (if any)
active=$(systemctl is-active sinkhole.service 2> /dev/null)
if [ "$active" = "active" ]; then
  echo "stopping existing sinkhole.service..."
  systemctl stop sinkhole.service
fi

enabled=$(systemctl is-enabled sinkhole.service 2> /dev/null)
if [ "$enabled" = "enabled" ]; then
  echo "disabling existing sinkhole.service..."
  systemctl disable sinkhole.service
  systemctl daemon-reload
fi

rm /etc/systemd/system/sinkhole.service
