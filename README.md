streamtools
===========

Streamtools is a tool for investigating and 
working with streams of data. Streamtools provides both a set of
useful stream processing patterns and the mechanism by which the patterns are 
combined together to create fully fledged systems. 

Patterns are implemented as *blocks* which can be connected together. Data, in
the form of a stream of discrete JSON objects, flow through connections and are operated on by
blocks. Depending on your resources, streamtools is capable of dealing with streams of data that operate up to thousands of JSON objects per second.

Blocks fall roughly into one of four categories:
* *source* blocks : collect data from the world.
* *sink* blocks : send data back out to the world.
* *transfer* blocks : receive messages, operate on them, and
  broadcast the result. 
* *state* blocks : receive incoming messages and learn from them,
  maintaining a state.

usage
=====

Run the service using `st`.

* Create blocks using `/create`
* Connect blocks using `/connect`
* Delete blocks using `/delete`
* Access blocks using `/blocks/`
* See all active blocks using `/list`
* See all available blocks using `/library`
