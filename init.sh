#!/bin/sh
nohup rclone rcd --rc-no-auth >/dev/null 2>&1 &
./BiliBackup -f 12345677 -pn 4 -remote onedrive:/test