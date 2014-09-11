/*

Package contentstore is for storing small to medium number of content-addresable
blobs ([]byte).

You can store a []blob of data in the store and retrieve it back using the
returned unique id (which is a sha1 of the content, but that's implementation
detail).


    // Note: in production code, check the error codes!
    store, _ := contentstore.New("mystore")
    id, _ := store.Put([]byte("my piece of content"))
    v, _ := store.Get(id)
    store.Close()

The data is de-duplicated (i.e. storing the same blob for the second time is
a no-op).

You cannot delete the data. It would be possible to implement but I haven't.

Where and why should you use content store?

Imagine any kind of content management system (a blog, a wiki, a forum). It needs
to store user posts, maybe uploaded images etc. Each piece of content is a blob.

You could store blobs in the database but it's much better to store them
in the filesystem. It's fast, simple and therefore reliable.

The simplest implementation would be to store each blob in a single file and
using e.g. sha1 of the content as file name. It gives you content de-duplication
for free.

It works but creates a lot of files which doesn't scale well past relatively
small (thousands) number of blobs. The filesystem has no problem storing
even millions of files (if you partition them intelligently into directories)
but e.g. a simple backup to s3 requires re-scanning all the files and
re-checking if they exist in s3. It's not the fastest process.

This package improves on that by storing multiple blobs in a single, append-only
file. This significantly reduces number of files created (and you can control
how big a single file can get).

It's sheer elegance in its simplicity.

*/
package contentstore
