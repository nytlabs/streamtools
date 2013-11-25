streamtools
===========

Streamtools is a [pattern language](http://en.wikipedia.org/wiki/Pattern_language) 
for working with streams of data. Streamtools provides both a set of
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


philosophy
==========

Streamtools is the result of three sets of observations we've been making in the
NYT R&D lab for a while now. These are:
* *The world is nonstationary.* In the past our analysis was typically based on a snapshot of
  available data, and we would forget, for a moment, that the
  underlying system is constantly changing. We call this assumption the
  '[stationarity](http://en.wikipedia.org/wiki/Stationary_process)' assumption, and it's a pretty common one. However, as far as assumptions go, it's not a particularly needed one, and can lead to some poor
  modelling. 
  blind to 
* *All data starts off life as a stream.* Every sensor collects data
  sequentially, be it a microarray, a thermometer, or a grad student with a
  clip-board. We take these streams of data and 'tabulate' them, to make them
  easier to store and work with in the future. However, as "big data" gives way
  to "data streams": endless amounts of data that arrive via the network, we are
  reminded that tabulation is a very specific act, and often neither necessary
  or desired. By drawing on signal processing and control systems tools, we can
  learn from streams of data as they arrive at our computer.
* *Creative abduction.* A distinguishing concept of data science is that it
  allows, and promotes, a playful approach to data analysis and modelling. Known
  as [abduction](http://en.wikipedia.org/wiki/Abductive_reasoning), this form of
  reasoning allows the scientist to reason from data to hypothesis, rather than
  the more traditional hypothesis testing framework. However, tools for
  performing this sort of reasoning are almost non-existent, which seems like a
  shame!

Streamtools, therefore, aims to provide a set of tools that are inherently
non-stationary, that deal natively with streams of data, and that allow playful
interactions with the data. 

usage
=====

Run the service using `st`.

* Create blocks using `/create`
* Connect blocks using `/connect`
* Delete blocks using `/delete`
* Access blocks using `/blocks/`
* See all active blocks using `/list`
* See all available blocks using `/library`
