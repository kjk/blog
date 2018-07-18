#!/bin/bash
set -u -e -o pipefail -o verbose

go build -o blog
./blog -deploy
./netlifyctl -A $NETLIFY_TOKEN deploy || true
cat netlifyctl-debug.log || true
