STREAMTOOLS - Timing
====================

Tools to adjust the timing of the stream. 

* `synchronize` - create a lagged stream that respects the internal timestamp of the stream's messages.
* `metronome` - create a stream of grouped messages, emitted at a regular interval 
* `reduce_by_keyvalue` - create a stream of grouped messages, grouped by value of a key, and emitted after not seeing that value after a fixed period
