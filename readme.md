[![Build Status](https://travis-ci.org/kjk/blog.svg?branch=master)](https://travis-ci.org/kjk/blog)

This is a Go program that generates my website/blog https://blog.kowalczyk.info.

I use [Notion](https://notion.so) to write most of the content.

This custom Go program downloads pages from Notion, caches it in `notion_cache` directory, converts to static HTML files and deploys to [Netlify](https://www.netlify.com/).

To extract my content from Notion I [reverse engineered their API](https://blog.kowalczyk.info/article/88aee8f43620471aa9dbcad28368174c/how-i-reverse-engineered-notion-api.html) and wrote a [Go library](https://github.com/kjk/notionapi).

I wrote an article about [how I got to this point](https://blog.kowalczyk.info/article/a8cf04d756ec4963905960822b004440/powering-a-blog-with-notion-and-netlify.html).
