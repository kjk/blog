#!/bin/bash
set -u -e -o pipefail

if [ -z ${1+x} ]; then
    echo "must provide page id"
    exit 1
fi

if [ "$1" == "" ]; then
    echo "must provide page id"
    exit 1
fi

go build -o blog
./blog -redownload-page $1
