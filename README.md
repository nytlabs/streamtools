STREAMTOOLS
===========

A set of tools for dealing with streams of JSON, using NSQ as a backend. The tools are envisaged as a set of binaries, not unlike the standard \*NIX tools, except with NSQ (a distributed queueing system) in place of unix pipes. Each tool is thought of as a block and can be connected together with other blocks to form complex data processing systems.


How to build a new block
========================

There are four types of blocks, each represented as an interface:

* InBlock - only have inbound connections (i.e. a sink block)
* OutBlock - only have outbound connections (i.e. a source block)
* InOutBlock - have both inbound and outbound connections (i.e. a transform block)
* StateBlock - has an internal state (i.e. min, max etc)

To make a new block, implement the appropriate interface in the `streamtools` package, then make a binary in `blocks/yourBlock/`. 
