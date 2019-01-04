Title: Extreme (size) optimization in C, C++ and Objective-C.

Extreme (size) optimization in C, C++ and Objective-C.
======================================================

by [Krzysztof Kowalczyk][]

What is this about?
-------------------

This document presents extreme techniques for optimizing the size of C,
C++ and Objective-C programs. It’s about optimizing both the size of the
final executable and the amount of memory used at runtime.

![][]:http://www.flickr.com/photos/strict/512024537/

Sometimes the code will also become faster, although speed is not the
main focus.

All examples come from optimizing real-life code.

Table of contents
-----------------

1.  Why optimize?
2.  [An optimization story][]
3.  Things you need to know
4.  Measuring size
5.  Asking the compiler to optimize for you
6.  Optimizing size of structures
7.  Optimizing arrays of structures
8.  Passing by value vs. passing by reference
9.  Allocating without freeing
10. Smaller code through better architecture
11. [Case study: optimizing disassembler in v8][]
12. Other misc ideas

  [Krzysztof Kowalczyk]: https://blog.kowalczyk.info
  []: >https://farm1.static.flickr.com/206/512024537_3da49317eb_m.jpg
  [An optimization story]: optimization_story.html
  [Case study: optimizing disassembler in v8]: optimizing_v8.html
