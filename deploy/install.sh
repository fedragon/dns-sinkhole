#!/usr/bin/env sh

# clean up old sinkhole service (if any)
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

# install new sinkhole service
mkdir -p sink/bin
mv hosts sink/
mv hole sink/bin/

mv sinkhole.service /etc/systemd/system/

echo "enabling sinkhole.service..."
systemctl daemon-reload
systemctl enable sinkhole.service

echo "run 'sudo systemctl start sinkhole.service' to start sinkhole"
