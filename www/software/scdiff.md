Title: scdiff

# What is scdiff?

Imagine this: you've just made changes to code kept in
[CVS](http://www.cvshome.org/), [Subversion](http://subversion.tigris.org/) or
[Git](http://git-scm.com/) repository. You're ready to check them in but you
want to take a last look at the changes. Usually you would do `cvs diff -u` or
`svn diff`. This program allows you to see the changes with an external gui
diff program. I find it much easier to understand the changes that way (as
opposed to looking at unified diff in the console). By default it uses
`windiff.exe` but you can use `-diff` option to select any other program (e.g.
[WinMerge](http://winmerge.org) or [Araxis
Merge](http://www.araxis.com/index.html)) . Clearly, it's a tool for
developers.

### Usage

`scdiff [-h] [-old] [-cvs cvsCommand] [-cvsargs cvsOptions] [-diff
diffProgramPath]`

If you run scdiff without any arguments, it'll determine if a given directory
is under CVS or Subversion control, check for locally modified files and
launch external diff program showing local modifications. By default it uses
windiff (assumes that windiff.exe is in the `%PATH%`) but you can use
**`-diff`** option to use any other diff program that can be launched from
command lind. First two arguments given to the diff program are directories to
diff. This works for all diff programs I've tested it (windiff, WinMerge and
Araxis Merge).

By its nature (see how it works for more explanation) scdiff uses temporary
directory for storing original and modified files so even after you finish,
you can still see the result of previous diff. Option `-old` does exactly
that. It saves the time (getting files from repository may take some time).

To see built-in help, use `-h` option.

Option `-cvs` defaults to "cvs -z3". Option `-cvsargs` defaults to "-u -N". In
theory you shouldn't need to change them.

### Download

Download [scdiff.exe](https://kjkpub.s3.amazonaws.com/files/scdiff.exe).
Requires .NET Framework 2.0.

### Source code

You can get the sources from [project
site](http://code.google.com/p/kjk/source/browse/#svn/trunk/vctools/scdiff).

### Version history

0.5 (2009-03-11):

  * added Git support

0.4 (2004-12-08):

  * add `-cvs` parameter to provide the name of cvs executable. Default is "cvs -z3".
  * add `-cvsargs` parameter to provide additional args to cvs. Default is "-u -N"
  * fixes for handling new and deleted files in cvs  All 0.4 changes provided by [dB](http://www.dblock.org).

0.3 (2004-10-02):

  * [Eli Tucker](http://nerdmonkey.com) fixed a bug handling directories in cvs
  * no longer crash when handling `svn remove`. Still, handling of deleting files is far from perfect (currently they are silently ignored while we should show apropriate diff)
  * properly handle file names with spaces. Fix suggested by Raman Gupta

0.2 (2004-06-08):

  * now also shows files that only exists locally and are not present in `cvs` repository. If that bothers you, learn how to use `.cvsignore`
  * show version number when displaying help

0.1 (2004-06-03):

  * first version

### How it works

Not that it's terribly interesting, but just in case you were wondering.
First, we capture the output of `cvs diff -u` or `svn diff`. From that we
extract names of the files that are locally modified and the revision number
of the file before modifications. The we check out the originals (using `cvs
update -p -r rev` or svn cat ...), copy the originals to
`$tempDir/sc_originals`, our locally modified copies to `$tempDir/sc_altered`
and launch external diff program with `$tempDir/sc_originals` and
`$tempDir/sc_altered` as arguments. Pretty simple and possibly suboptimal
(subversion can do a diff without contacting remote repository, so it should
be possible to significantly speed up the program if I knew how to get the
original without asking remote repository).

### Todo

It's really a quick & dirty program, so there's potential for a lot of stuff
to be done. In the "blue sky" departement, I would like to have a full-fledged
program for browsing changes in CVS or Subversion repositories. And no, it's
not about re-writing [WinCVS](http://www.wincvs.org/) and the like for the fun
of it. WinCVS does much more that what I need my ideal program to do, but it
also doesn't do what I want (easily end efficiently browse changes).

But that's unlikely to happen, so here's a couple of things that could be
fixed:

  * as noted in "how it works", it should be possible to make Subversion case work without contacting remote repository
  * currently you can't launch two copies at the same time because they use the same temp directory for storing files so the second copy will be unable to access temp directory
  * currently windiff.exe must be in path. Could try to auto-detect full path by checking known paths or maybe looking in the registry
  * auto-detect other diff programs like WinMerge and Araxis Merge
  * downloading revisions from the repository takes time. Revisions do not change so they are perfect target for caching. We could locally cache revisions we retrieved so far to speed up a case of comparing local changes with the same revision multiple times. It happens quite often (we often get into develop/compare/fix/compare/fix/compare... cycle)

### Links

  * [CVS](http://cvshome.org) and [Subversion](http://subversion.tigris.org/) are source control systems. Use Subversion if you have a choice
  * [Windiff](http://msdn.microsoft.com/library/default.asp?url=/library/en-us/tools/tools/windiff.asp), [WinMerge](http://winmerge.org), [Araxis Merge](http://www.araxis.com/index.html) are diff/merge programs. [There are others](http://keithdevens.com/downloads#diff).
  * [WinCVS](http://www.wincvs.org/) is a GUI for managing CVS. It doesn't do what I need.

### Feedback

As noted, this program is simple, does one thing that is useful to me. It
might never get any better and there's not much to talk about. If you,
however, have a burning desire to talk to me about it (you know, comments, bug
reports, suggestions etc.), you can always [send me an e-mail](/).

