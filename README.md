streamtools
===========

Streamtools is a tool for investigating and 
working with streams of data. Streamtools provides both a set of
useful stream processing patterns and the mechanism by which the patterns are 
combined together to create fully fledged systems. 

Patterns are implemented as *blocks* which can be connected together. Data, in
the form of a stream of discrete JSON objects, flow through connections and are operated on by
blocks. Depending on your resources, streamtools is capable of dealing with streams of data that operate at up to thousands of JSON objects per second.

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
* *The world is nonstationary.* It's not uncommong for our analysis to be based on a snapshot of
  available data, arranged in a table. We forget, for a moment, that the
  underlying system is constantly changing and assume, instead, that the system's properties are static 
  during the period the data was collected. We call this assumption the
  '[stationarity](http://en.wikipedia.org/wiki/Stationary_process)' assumption, 
  and it's a pretty common one. However, when studying systems whose properties change over time, 
  this assumption can lead to some poor
  modelling. For example, when studying a system as complex as an audience, it is critical
  to realise that the audience's properties are changing on many different time
  frames. 
* *All data starts off life as a stream.* Every sensor collects data
  sequentially, be it a set of microarray, a thermometer, or a grad student with a
  clip-board. We take these streams of data and 'tabulate' them, to make them
  easier to store and work with in the future. However, as "big data" gives way
  to "data streams": endless amounts of data that arrive via the network, we are
  reminded that tabulation is a very specific act, and often neither necessary
  or desired. By drawing on signal processing and control systems tools, we can
  learn from streams of data as they arrive at our computer.
* *Creative abduction.* A distinguishing concept of data science is that it
  allows, and promotes, a playful approach to data analysis and modelling. Known
  as [abduction](http://en.wikipedia.org/wiki/Abductive_reasoning), this form of
  reasoning allows the scientist to reason their way from data to hypothesis, rather than
  the more traditional hypothesis testing framework. However, tools for
  performing this sort of reasoning are almost non-existent, which seems like a
  shame!

Streamtools, therefore, aims to provide a set of tools that are inherently
non-stationary, that deal natively with streams of data, and that allow playful
interactions with the data. 

what streamtools isn't
======================

Streamtools is definitely NOT a production quality queueing system. For these,
see:
* [NSQ](http://bitly.github.io/nsq/)
* [RabbitMQ](http://www.rabbitmq.com/)
* [ZeroMQ](http://zeromq.org/)
* [AmazonSQS](http://aws.amazon.com/sqs/)

Streamtools also isn't a production quality event processing framework. Streamtools provides a set
of patterns for performing analysis on streams of data - the user should not
expect to write any code at all. If you're looking for an event processing
framework for building production systems, check out the following:
* [Storm](http://storm-project.net/)
* [Spark](http://spark.incubator.apache.org/)
* [Amazon Kinesis](http://aws.amazon.com/kinesis/)

We expect streamtools to be very useful for prototyping systems that
subsequently rely on these technologies!

usage
=====

Run the service using `st`.

* Create blocks using `/create`
* Connect blocks using `/connect`
* Delete blocks using `/delete`
* Access blocks using `/blocks/`
* See all active blocks using `/list`
* See all available blocks using `/library`
