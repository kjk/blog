[![Build Status](https://travis-ci.org/kjk/blog.svg?branch=master)](https://travis-ci.org/kjk/blog)

This is a Go program that generates my website/blog https://blog.kowalczyk.info.

I use [Notion](https://notion.so) to write most of the content.

This custom Go program downloads pages from Notion, caches it in `notion_cache` directory, converts to static HTML files and deploys to [Netlify](https://www.netlify.com/).

To extract my content from Notion I [reverse engineered their API](https://blog.kowalczyk.info/article/88aee8f43620471aa9dbcad28368174c/how-i-reverse-engineered-notion-api.html) and wrote a [Go library](https://github.com/kjk/notionapi).

I wrote an article about [how I got to this point](https://blog.kowalczyk.info/article/a8cf04d756ec4963905960822b004440/powering-a-blog-with-notion-and-netlify.html).

### Build and run

Note that this blog is very specific to my i.e. it has a collection of static pages in .html and pulls data from my Notion pages).

If you want to adopt for your purposes, here are the things you might want to change:
* `analyticsCode` in `main.go`
* `notionBlogsStartPage` in `articles.go`. this is a page that has a list of blog articles, which are treated specially (they form the blog part)
* `notionWebsiteStartPage` in `articles.go` this is a page for the root of the website's content
* `notionGoCookbookStartPage` in `articles.go` - well, this and all code related to it should be removed. This is a page for the root of my "Go Cookbook" mini-book
* make those pages public (but disable search text indexing) (via `Share` button in Notion, at the top right).

Then you can see `s\preview.ps1` script to see what the build process is, which currently is:

1. `go build`
2. Get the `token_v2` from the cookies of a browser logged into Notion.so
3. Set environment variable, e.g. for bash: `NOTION_TOKEN=<value of token_v2>`
4. run `./blog` or `./blog -preview` to also start a local web server for previewing files
5. Now you can start a local webserver in the `netlify_static` directory (e.g. `npx live-server netlify_static`)

HTML files are generated in `netlify_static` directory because I deploy to Netlify but since it's mostly a static website, you can deploy it pretty much anywhere.
