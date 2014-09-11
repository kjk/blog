contentstore
============

Go library for storing sha1-addressable blobs ([]byte slices)

A sample use:

    // Note: in production code, check the error codes!
    store, _ := contentstore.New("mystore")
    id, _ := store.Put([]byte("my piece of content"))
    v, _ := store.Get(id)
    store.Close()

Documentation: http://godoc.org/github.com/kjk/contentstore

[![Build Status](https://drone.io/github.com/kjk/contentstore/status.png)](https://drone.io/github.com/kjk/contentstore/latest)
