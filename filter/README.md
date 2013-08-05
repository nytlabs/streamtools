STREAMTOOLS - Filter
=====================

Filter a stream in various manners

* `demux_by_key` - reads a topic, splits the stream into two - one that contains the specified key and one that doesn't
* `filter_by_key` - throws out any message that does (or doesn't - configurable) contain key
* `filter_by_keyvalue` - throws out any message that doesn't satisfy the given value comparison
* `join_by_key` - reads two topics, a primary and secondary, waits for `timeout` seconds, joining and emitting messages if a match shows up, otherwise discards secondary topic message and emits primary topic message. 
