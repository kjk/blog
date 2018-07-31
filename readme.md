[![Build Status](https://travis-ci.org/kjk/blog.svg?branch=master)](https://travis-ci.org/kjk/blog)

This is a static blog for https://blog.kowalczyk.info.

The content is managed in https://notion.so, cached in `notion_cache/` directory.

This custom program converts content from Notion to static html and deploys to https://www.netlify.com/.

To get data from Notion I [reverse engineered their API](https://blog.kowalczyk.info/article/88aee8f43620471aa9dbcad28368174c/how-i-reverse-engineered-notion-api.html) and wrote a [Go libray](https://github.com/kjk/notionapi).

I wrote an article about [how I got to this point](https://blog.kowalczyk.info/article/a8cf04d756ec4963905960822b004440/powering-a-blog-with-notion-and-netlify.html).
