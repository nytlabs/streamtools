streamtools
===========

Streamtools is a [pattern language](http://en.wikipedia.org/wiki/Pattern_language) 
used for dealing with streams of data. Streamtools provides both a set of
useful stream processing patterns and the mechanism by which the patterns are 
combined together to create fully fledged systems. 

Patterns are implemented as *blocks* which can be connected together. Data, in
the form of JSON objects, flow through connections and are operated on by
blocks. 

Blocks fall roughly into one of four categories:
* *source* blocks : these collect data from the world.
* *sink blocks* : these send data back out to the world, e.g. by
  querying an API, writing to a databse, or broadcasting on a websocket.
* *transfer blocks* : these recieve messages, operate on them, and
  broadcast the result. 
* *state blocks* : these recieve incoming messages and learn from them,
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
