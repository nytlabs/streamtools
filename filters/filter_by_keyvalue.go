package main

import (
	flag
)

var (
	key = flag.String("key", "", "key against which to filter")
	value_equals = flag.String("value", "", "value for comparison")
	comparator = flag.String("comparator", "equal", "type of comparison")
)
