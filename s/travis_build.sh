#!/bin/bash
set -u -e -o pipefail -o verbose

build()
{
    echo "building"
    go build -o blog
    ./blog -deploy
    ./netlifyctl -A $NETLIFY_TOKEN deploy || true
    cat netlifyctl-debug.log || true
}

setup_git()
{
  git config --global user.email "kkowalczyk@gmail.com"
  git config --global user.name "Krzysztof Kowalczyk"
  git config --global github.user "kjk"
  git config --global github.token "${GH_TOKEN}"
}

update_from_notion()
{
    echo "cron: updating from notion"
    setup_git()

    go build -o blog
    ./blog -redownload-notion
    git commit -am "travis: update from notion"
    git push
}

if [ "${TRAVIS_EVENT_TYPE}" == "cron"]; then
    update_from_notio
else
    build
fi
