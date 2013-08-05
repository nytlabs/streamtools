stream_tools
============

Tools for working with streams of data. Heavily using golang and [NSQ](https://github.com/bitly/nsq).

Tools are currently organized in the following categories:

#### IMPORT
Creates a stream.

_e.g. `poll_to_streamtools`, `csv_to_streamtools`, `nsq_to_streamtools`, etc._

#### EXPORT
Non-stream output.

_e.g. `streamtools_to_vega`, `streamtools_to_ws`, etc._

#### TIMING
Takes stream input, and outputs stream with modified time-order of messages.

_e.g. `synchronizer`, `metronome`, etc._

#### FILTER
Takes stream input, and outputs stream filtered by conditions.

_e.g. `filter_by_key`, `sampler`, `rate_limiter`, etc._

#### TRANSFORM
Takes stream input, and outputs new stream of messages. Single-in-single-out.

_e.g. `feature_extractor`, `type_inferencer`, etc._

#### TRACKING
Takes stream input, maintains state that is continuously updated on each input message. No output.

_e.g. `distribution`, `boundedness`, etc._