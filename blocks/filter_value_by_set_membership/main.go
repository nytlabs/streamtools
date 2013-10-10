package main

import (
    "flag"
    "github.com/bitly/go-simplejson"
    "github.com/nytlabs/streamtools/streamtools"
    "log"
)

var (
    readTopic  = flag.String("read-topic", "", "topic to read from")
    writeTopic = flag.String("write-topic", "", "topic to write to")
    path        = flag.String("path", "", "key agsinst which to filter in dot notation (key_a.key_b.key_c)")
    set     = flag.String("set", "", "json array of values to pass filter ('[\"red\",\"blue\",\"green\"]')")
    name       = flag.String("name", "filter_value_by_set_membership", "name of block")
)

func main() {
    flag.Parse()
    streamtools.SetupLogger(name)

    log.Println("reading from", *readTopic)
    log.Println("writing to", *writeTopic)
    log.Println("path ", *path)
    log.Println("set ", *set)

    block := streamtools.NewTransferBlock(streamtools.FilterValueBySetMembership, *name)
    rule, err := simplejson.NewJson([]byte("{}"))
    if err != nil {
        log.Fatal(err.Error())
    }

    //setJson, _ := simplejson.NewJson([]byte(*set))
    rule.Set("path", *path)
    rule.Set("set", *set)

    block.RuleChan <- rule
    block.Run(*readTopic, *writeTopic, "8080")
}
