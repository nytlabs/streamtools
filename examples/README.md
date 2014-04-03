# Examples

See the wiki for [instructions](https://github.com/nytlabs/streamtools/wiki/examples) on how to run these examples.

## use cases

* citibike.json : poll the citibike API! This pattern keeps track of how many bikes
  are avaialable outside the New York Times HQ.
* 1-usa-gov.json : listen to the 1.usa.gov long lived HTTP stream
* wikipedia-edits.json : track wikipedia editors as they edit wikipedia live

## components

* poller.json : poll an HTTP endpoint
* random-numbers.json : generate a sequence of random numbers
* check-join-filter.json : check if a message contains a value in a set and emit if so
* count-clear-filter.json : emit every N messages (aka integrate and fire
  neuron!)

## demos

These examples incorporate other technologies to describe systems where streamtools plays an integral part. 

* phoneExample/ : use streamtools to listen to the orientation collected from smartphones and emit the average of the orientation to build a simple crowd-voting mechanism. 
