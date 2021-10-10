#!/bin/bash

# jetson端末再起動時にwindowsサーバのフォルダーをマウントする
find /mnt/windows -mindepth 1 -maxdepth 1 -type d | while read path; do
  rm -rf "$path"
done

smbnetfs /mnt/windows
