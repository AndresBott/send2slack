#!/usr/bin/env sh

USERNAME="send2slack"

## add user
adduser --system --no-create-home $USERNAME


## change config files permissions
chown $USERNAME:root /etc/send2slack/server.yaml
chmod 600 /etc/send2slack/server.yaml
chmod 644 /etc/send2slack/client.yaml

## Enable the service
systemctl enable send2slack.service
systemctl start send2slack.service