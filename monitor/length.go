package main

import (
    "flag"
)

var(
    key = flag.String("key", "", "find the length of this key's array")   
    readTopic = flag.String("read_topic", "", "topic to read from")
    writeTopic = flag.String("write_topic", "", "topic to write to")
)

