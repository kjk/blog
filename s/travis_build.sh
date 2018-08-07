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

# this downloads the latest version of content from notion and checks it in
# this is triggered via daily cron builds on travis (they run at midnight)
update_from_notion()
{
    echo "cron: updating from notion"
    rm -rf netlify*
    setup_git
    git checkout master

    go build -o blog
    ./blog -redownload-notion

    echo "after build"
    git status
    git checkout Gopkg.lock
    git checkout netlify.toml
    git status
    git add notion_cache/*
    echo "after git add"
    git status
    now=`date +%Y-%m-%d %a`

    # "git commit" returns 1 if there's nothing to commit, so don't report this as failed build
    set +e
    git commit -am "travis: update from notion on ${now}"
    if [ "$?" -ne "0" ]; then
        echo "nothing to commit"
        exit 0
    fi
    set -e
    git push "https://${GH_TOKEN}@github.com/kjk/blog.git" master || true
}

if [ "${TRAVIS_EVENT_TYPE}" == "cron" ]; then
    update_from_notion
else
    build
fi
