STREAMTOOLS
===========

A set of tools for dealing with streams of JSON, using NSQ as a backend.

How to build a new block
========================

There are four types of blocks, each represented as an interface:

* InBlock - only have inbound connections (i.e. a sink block)
* OutBlock - only have outbound connections (i.e. a source block)
* InOutBlock - have both inbound and outbound connections (i.e. a transform block)
* StateBlock - has an internal state (i.e. min, max etc)

To make a new block, implement the appropriate interface in the `streamtools` package, then make a binary in `blocks/yourBlock/`. 
