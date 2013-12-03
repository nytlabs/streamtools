# StreamTools makefile

# requires system environment variable $GOPATH to be set, with a standard GoLang heirarchy, and to be run from the streamtools project directory


# NOTE: not sure if or how to abstract this
LOCAL_OS_LIB_SUBDIR=darwin_386

GOGET=go get
GOINSTALL=go install
BUILDFLAGS=
GOBIN=$(GOPATH)bin
GOPKGBITLY=$(GOPATH)pkg/$(LOCAL_OS_LIB_SUBDIR)/github.com/bitly
GOSRCBITLY=$(GOPATH)src/github.com/bitly
GOPKGGOOG=$(GOPATH)pkg/$(LOCAL_OS_LIB_SUBDIR)/code.google.com/p/go.net
GOSRCGOOG=$(GOPATH)src/code.google.com/p/go.net

ALL_PACKAGES=http_to_streamtools random_stream streamtools_to_pubsub streamtools_to_ws\
 mask metronome synchronizer array_length type_inferencer
	# csv_to_streamtools nsq_to_streamtools demux_by_key filter_by_keyvalue join_by_key reduce_by_keyvalue boundedness distribution

# add all executable and library src & targets to BASEPATH
BASEPATH=$(GOBIN):$(GOPKGBITLY):$(GOPKGBITLY)/nsq:$(GOSRCBITLY)/go-simplejson:$(GOSRCBITLY)/nsq/nsq:$(GOPKGGOOG):$(GOSRCGOOG)/websocket
# add BASEPATH and all source directories to VPATH
VPATH=$(BASEPATH)\
	:import:import/http_to_streamtools:import/random_stream\
	:export/streamtools_to_pubsub:export/streamtools_to_ws\
	:filter/mask\
	:timing/metronome:timing/synchronizer\
	:transform/array_length\
	:transform/type_inferencer

all: $(ALL_PACKAGES)

# libraries
dependencies: nsq.a go-simplejson.a websocket.a

nsq.a: command_test.go

command_test.go:
	$(GOGET) github.com/bitly/nsq

# .PHONY: nsq_0221
# nsq_0221:
# 	$(GOGET) github.com/bitly/nsq/tree/v0.2.21
# 	# github.com/bitly/nsq/tree/v0.2.21/nsq

go-simplejson.a: simplejson.go

simplejson.go:
	$(GOGET) github.com/bitly/go-simplejson

websocket.a: websocket.go

websocket.go:
	$(GOGET) code.google.com/p/go.net/websocket


# streams
http_to_streamtools: nsq.a http_to_streamtools.go
	$(GOINSTALL) $(BUILDFLAGS) ./import/http_to_streamtools

random_stream: go-simplejson.a random_stream.go
	$(GOINSTALL) $(BUILDFLAGS) ./import/random_stream

# csv_to_streamtools: csv_to_streamtools.go
# 	go build -o $(GOBIN)/csv_to_streamtools $(BUILDFLAGS) ./import/csv_to_streamtools.go

# nsq_to_streamtools: nsq_to_streamtools.go
# 	$(GOINSTALL) $(BUILDFLAGS) ./import/nsq_to_streamtools.go

streamtools_to_pubsub: nsq.a streamtools_to_pubsub.go
	$(GOINSTALL) $(BUILDFLAGS) ./export/streamtools_to_pubsub

streamtools_to_ws: nsq.a websocket.a streamtools_to_ws.go
	# errors
	# $(GOINSTALL) $(BUILDFLAGS) ./export/streamtools_to_ws

mask: mask.go
	# errors
	# $(GOINSTALL) $(BUILDFLAGS) ./filter/mask

# demux_by_key: demux_by_key.go
# 	$(GOINSTALL) $(BUILDFLAGS) ./filter/demux_by_key.go

# filter_by_keyvalue: filter_by_keyvalue.go
# 	$(GOINSTALL) $(BUILDFLAGS) ./filter/filter_by_keyvalue.go

# join_by_key: join_by_key.go
# 	$(GOINSTALL) $(BUILDFLAGS) ./filter/join_by_key.go

# .PHONY: filter
# filter: demux_by_key.go filter_by_keyvalue.go join_by_key.go
# 	$(GOINSTALL) $(BUILDFLAGS) ./filter/demux_by_key.go ./filter/filter_by_keyvalue.go ./filter/join_by_key.go

metronome: nsq.a metronome.go
	$(GOINSTALL) $(BUILDFLAGS) ./timing/metronome

synchronizer: go-simplejson.a nsq.a synchronizer.go prority_queue.go
	# need: nsq_0221
	# $(GOINSTALL) $(BUILDFLAGS) ./timing/synchronizer

# reduce_by_keyvalue: reduce_by_keyvalue.go
# 	$(GOINSTALL) $(BUILDFLAGS) ./timing/reduce_by_keyvalue.go

# boundedness: boundedness.go
# 	$(GOINSTALL) $(BUILDFLAGS) ./tracking/boundedness.go

# distribution: distribution.go
# 	$(GOINSTALL) $(BUILDFLAGS) ./tracking/distribution.go

array_length: go-simplejson.a nsq.a array_length.go
	$(GOINSTALL) $(BUILDFLAGS) ./transform/array_length

type_inferencer: go-simplejson.a nsq.a type_inferencer.go
	$(GOINSTALL) $(BUILDFLAGS) ./transform/type_inferencer


.PHONY: clean
clean:
	for i in $(ALL_PACKAGES); do rm -f $(GOBIN)/$$i; done;

