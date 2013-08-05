stream_tools
============

Tools for working with streams of data. Using [NSQ](https://github.com/bitly/nsq).

Tools are currently organized in the following categories:

#### IMPORT
: Creates a stream.

e.g. `poll_to_streamtools`, `csv_to_streamtools`, `nsq_to_streamtools`, etc.

#### EXPORT
: Non-stream output.

e.g. `streamtools_to_vega`, `streamtools_to_ws`, etc.

#### TIMING
: Takes stream input, and outputs stream with modified time-order of messages.

e.g. `synchronizer`, `metronome`, etc.

#### FILTER
: Takes stream input, and outputs stream filtered by conditions.

e.g. `filter_by_key`, `sampler`, `rate_limiter`, etc.


#### TRANSFORM
: Takes stream input, and outputs new stream of messages. Single-in-single-out.

e.g. `feature_extractor`, `type_inferencer`, etc.

#### TRACKING
: Takes stream input, maintains state that is continuously updated on each input message. No output.

e.g. `distribution`, `boundedness`, etc.