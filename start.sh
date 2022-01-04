#!/bin/bash
# shellcheck disable=SC2009
SYSTEM=$(uname -s)                    #获取操作系统类型
if [ "$SYSTEM" = "Linux" ] ; then
pid=$(ps -ef | grep dockerd | grep -v grep | awk -F" " '{print $2}')
if [ ${#pid} -ge 0 ]; then
    kill -9 "$pid"
fi
dockerd&
nohup ./translate-server &
fi



