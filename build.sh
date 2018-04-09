#!/usr/bin/env bash - 
#===============================================================================
#
#          FILE: build.sh
# 
#         USAGE: ./build.sh 
# 
#   DESCRIPTION: 
# 
#       OPTIONS: ---
#  REQUIREMENTS: ---
#          BUGS: ---
#         NOTES: ---
#        AUTHOR: YOUR NAME (), 
#  ORGANIZATION: 
#       CREATED: 09.04.2018 15:08
#      REVISION:  ---
#===============================================================================

set -o nounset                              # Treat unset variables as an error
for i in $#
do
    if [[ "$i" == "1" && "${!i}" == "ok" ]]
    then
        go build -o checkexchange main.go
        if [ $? == 0 ]
        then
            echo "build ok!"
        else
            echo "build error!"
        fi
    else
        go run main.go
    fi
done
