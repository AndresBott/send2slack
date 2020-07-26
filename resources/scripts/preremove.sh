#!/usr/bin/env sh

## stop the service
systemctl stop send2slack.service
systemctl disable send2slack.service

systemctl stop send2slack-mbox-watcher.service
systemctl disable send2slack-mbox-watcher.service