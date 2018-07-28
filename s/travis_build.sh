#!/bin/bash
set -u -e -o pipefail

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
    rm -rf netlify*
    setup_git
    git status
    git checkout -b master

    go build -o blog
    ./blog -redownload-notion

    git checkout Gopkg.lock
    git status
    git add notion_cache/*
    echo "after git add"
    git status
    git commit -am "travis: update from notion"
    git push "https://${GH_TOKEN}@github.com/kjk/blog.git" master
}

if [ "${TRAVIS_EVENT_TYPE}" == "cron" ]; then
    update_from_notion
else
    build
fi
