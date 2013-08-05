package main

import (
    "flag"
)

var (
    key          = flag.String("key", "", "key against which to filter")
    inTopic      = flag.String("in_topic", "", "topic to read from")
    outTopicPrefix     = flag.String("out_topic_prefix", "", "prefix of output stream")
)

