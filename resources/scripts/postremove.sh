#!/usr/bin/env sh
USERNAME="send2slack"

## try to remove the user
if [ -x "$(command -v deluser)" ]; then
   deluser --quiet --system $USERNAME > /dev/null || true
else
   echo >&2 "not removing $USERNAME system account because deluser command was not found"
fi

