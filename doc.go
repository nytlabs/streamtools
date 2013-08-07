// The MIT License (MIT)
//
// Copyright (c) 2013 The New York Times Company
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//
// <<< the Streamtools manifesto goes here >>>
//
// Tools for working with streams of data. Heavily using golang and [NSQ](https://github.com/bitly/nsq).
// Tools are currently organized in the following categories:
//
// IMPORT
// Creates a stream.
// e.g. `poll_to_streamtools`, `csv_to_streamtools`, `nsq_to_streamtools`, etc.
//
// EXPORT
// Non-stream output.
// e.g. `streamtools_to_vega`, `streamtools_to_ws`, etc.
//
// TIMING
// Takes stream input, and outputs stream with modified time-order of messages.
// e.g. `synchronizer`, `metronome`, etc.
//
// FILTER
// Takes stream input, and outputs stream filtered by conditions.
// e.g. `filter_by_key`, `sampler`, `rate_limiter`, etc.
//
// TRANSFORM
// Takes stream input, and outputs new stream of messages. Single-in-single-out.
// e.g. `feature_extractor`, `type_inferencer`, etc.
//
// TRACKING
// Takes stream input, maintains state that is continuously updated on each input message. No output.
// e.g. `distribution`, `boundedness`, etc.