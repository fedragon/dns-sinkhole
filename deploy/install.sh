#!/usr/bin/env sh
set -x

active=$(systemctl --user is-active sinkhole.service)
if [ "$active" = "active" ]; then
  systemctl --user stop sinkhole.service
fi

mkdir -p sink/bin
mv hosts sink/
mv hole sink/bin/

mkdir -p ~/.config/systemd/user
mv sinkhole.service ~/.config/systemd/user/
systemctl --user daemon-reload
enabled=$(systemctl --user is-enabled sinkhole.service)

if [ "$enabled" != "enabled" ]; then
  systemctl --user enable sinkhole.service
fi

echo "run 'systemctl --user start sinkhole.service' to start sinkhole"
