## Intro
Streamtools is a graphical toolkit for dealing with streams of data. Streamtools makes it easy to explore, analyse, modify and learn from streams of data.

You'll primarily interact with streamtools in the browser. However, since all functionality is exposed over HTTP, you can use tools like curl to send commands and even treat any part of streamtools as an API endpoint. More about that later.


## Install
### Binary

Download the appropriate version of the latest streamtools for your operating system from our [releases on github](https://github.com/nytlabs/streamtools/tags).

Extract the archive (it'll either be a .tar.gz or .zip).

Navigate to the extracted folder and run the ``st`` executable.

* from Terminal: ```./st```
* from Finder: double-click

You should see streamtools start up, telling you it's running on port 7070.

Now, open a browser window and point it at [localhost:7070](http://localhost:7070/). You should see a (nearly) blank page. At the bottom you should see a status bar that says `client: connected to Streamtools` followed by a version number. Congratulations! You're in.


### Source

Make sure you have go, git, hg, and bzr installed. You can download go for [Mac OS X](http://golang.org/doc/install#osx), [Linux and FreeBSD](http://golang.org/doc/install#tarball) and [Windows](http://golang.org/doc/install#windows) from the [golang.org](http://golang.org/) website. Git, hg and bzr are simple enough to install using homebrew, apt or your OS package manager of choice.

Once you have these dependencies, compile streamtools with these commands:

```
mkdir -p ~/go/src/github.com/nytlabs
cd ~/go/src/github.com/nytlabs
git clone git@github.com:nytlabs/streamtools.git
cd streamtools
make
```

To start the streamtools server:

```
$ ./build/st
Apr 30 19:44:36 [ SERVER ][ INFO ] "Starting Streamtools 0.2.5 on port 7070"
```

You should see a message similar to the one above letting you know streamtools is running at port 7070.


## Getting Started

Streamtools is a binary that can run on your local machine or a remote server. We usually run it using upstart on an ubuntu ec2 server in Amazon's cloud. To begin with, though, we'll assume that you're running streamtools locally, on a machine you can touch. We're also going to assume you're running OSX or Linux. If you're a Windows user, we do provide binaries, but don't know much about how to interact with a Windows machine - you will need to translate these instructions to Windows yourself.

Before we go any further, you should make sure you've installed streamtools. Check out the directions on starting the server, either from a binary release or from source, if you haven't already done so. 

You should see streamtools start up, telling you it's running on port 7070.

Now, open a browser window and point it at [localhost:7070](http://localhost:7070/). You should see a (nearly) blank page. At the bottom you should see a status bar that says `client: connected to Streamtools` followed by a version number. You're in!

As a "Hello World", try double-clicking anywhere on the page above the status bar, type `fromhttpstream` and hit enter. This will bring up your first block. Double-click on the block and enter `http://developer.usa.gov/1usagov` in the `Endpoint` text-box. Hit the update button. Now double-click on the page and make a `tolog` block. Finally, connect the two blocks together by first clicking on the `fromhttpstream` block's OUT route (a litle black square on the bottom of the block) to the `tolog` block's IN route (which is the little black square on the top of the block). Click on the status bar. After a moment, you should start to see JSON scroll through the log - these are live clicks on the US government short links! Click anywhere on the log to make it go away again. 

### How It Works

Streamtools' basic paradigm is straightforward: data flows from *blocks* through *connections* to other blocks. 

* A block perfoms some operation on each message it receives, and that operation is defined by the block's *type*. 
* Each block has zero or more *rules* that define that block's behaviour. 
* Each block has a set of named *routes* that can receive data, emit data, or respond to queries.
* You can connect blocks together, via their routes, using connections. You can connect to any inbound route, and so data flowing through streamtools can be used to set the rules of the blocks in the running pattern.
* We call a collection of connected blocks a *pattern*, and it is possible to export and import whole patterns from a running instance of streamtools. 

Together, these 5 concepts--blocks, rules, connections, routes and patterns--form the basic vocabulary we use to talk about streamtools, and about streaming data systems.


## Blocks

Each block is briefly detailed below, along with the rules that define each block. To make a block in streamtools, double-click anywhere on the page and type the name of the block as it appears below. For programmatic access, see the [API](#api) docs.

Blocks rely on some general concepts:

* _gojee path_: The path rules all use [gojee](https://github.com/nytlabs/gojee) syntax to specify which value you'd like to use in the block. Paths always start with the period, which indicates the top-level of the message. So if you want to refer to the whole message, use `.`. If you want to refer to a specific value, that value follows the first period. So if you have a message that looks like

```   
        {
            "user":{
                "username":"bob_the_user"
                "id": 1234
            }
        }
```

and you'd like to refer to the username, the gojee path would be `.user.username`.

* _gojee expression_: [gojee](https://github.com/nytlabs/gojee) also allows for expressions. So we can write expressions like `.user.id > 1230`, which are especially useful in the `filter` and `map` blocks.  
* _duration string_: We use Go's duration strings to specify time periods. They are a number followed by a unit and are pretty intuitive. So `10ms` is 10 milliseconds; `5h` is 5 hours and so on. 
* _route_: every block has a set of routes. Routes can either be inbound, query, or outbound routes. Inbound routes receive data from somewhere and send it to the block. Query routes are two-way: they accept an inbound query and return information back to the requester. Outbound routes send data from a block to a connection.

### Core

These blocks affect data once it's in streamtools. 

* **ticker**. This block emits the time regularly. It's useful for polling other blocks, like ```webRequest``` or ```count```. The time between emitting messages is specified by the `Interval`.
    * Rules:
        * `Interval`: duration string (`1s`)

* **bang**. Sometimes you just want to kick another block into action once, without waiting for some duration of time for the ```ticker``` block to kick it into action. The bang block is here for you.
	* Rules: none.

Connect the bang block's OUT endpoint to another block's IN endpoint. When you want to bang the connected block, click the bang block's query endpoint, i.e. the red square at its upper right corner.

You should see data start to flow further down the pattern of blocks.

* **javascript**. This block creates a Javascript VM and runs a bit of Javascript once per message. In order to get data in and out of Javascript, the block creates a global variable specified by `MessageIn` that contains the incoming message. Once the script is finished executing, the block takes the value from the global variable specified by `MessageOut`.
    * Rules:
        * `MessageIn`: string (`input`)
        * `MessageOut`: string (`output`)
        * `Script`: Javascript (`output = input`)
        
* **join**. This block joins two streams together. It waits until it has seen a message on both its inputs, then emits the joined message. 

* **map**. This block maps inbound data onto outbound data. The `Map` rule needs to be valid JSON, where each key is a string and each value is a valid [gojee](https://github.com/nytlabs/gojee) expression.
    * Rules:
        * `Map`: [gojee](https://github.com/nytlabs/gojee) expression
        * `Additive`: (`True`)

* **mask**. This block allows you to select a subset of the inbound message. To create a mask, you need to build up an empty JSON that looks like the message you'd like out. So, for example, if your inbound message looks like

        {
          "A":"foo",
          "B":"bar"
        }
    and you just want `B` to come out of the `mask` block, you'd make the `Mask` rule:

        {
          "B":{}
        }
    You can supply any valid JSON to the `Mask` block. If you specify an empty JSON `{}`, then all values will pass.
    * Rules:
        * `Mask`: mask JSON
        
* **filter**. The `filter` block applies the provided rule to incoming messages. If the rule evaluates to `true`, the messages is emitted. If the rule evaluates to `false`, the messages is discarded. The `Filter` rule can be any valid [gojee](https://github.com/nytlabs/gojee) expression. So, for example, if the inbound message looks like

        {
            "temperature": 43
        }
    and you only want to emit messages when the `temperature` value is above 50, then the `Filter` rule would be

        .temperature > 50

    * Rules:
        * `Filter`: [gojee](https://github.com/nytlabs/gojee) expression (`. != null`)

* **unpack**. The unpack block takes an array of objects and emits each object as a separate message. See the [citibike example](https://github.com/nytlabs/streamtools/blob/master/examples/citibike.json#L77), where we unpack a big array of citibike stations into individual messages we can filter.  
    * Rules:
        * `Path`: [gojee](https://github.com/nytlabs/gojee) path

* **sync**. The sync block takes a disordered stream and creates a properly timed, ordered stream at the expense of introducing a lag. To explain this block, imagine you have a stream that looks like

        {"val":"a", "time":23 } ... {"val":"b", "time":14} ... {"val":"c", "time":10}

    Ideally you'd like the stream to be ordered by the timestamp in the message, so the `c` message comes first, the `b` message comes second and the `a` message comes third. In addition, you'd like the time between the messages to respect the timestamp inside the message. 

    The sync block achieves this by storing the stream for a fixed amount time (the `Lag`) and then emitting at the time inside the inbound messages plus the lag. This means we have to wait for a while to get our messages, but when we do get them, they're in a stream whose dynamics reflect the timestamp inside the message. 

    This can be very helpful if the plumbing between your sensor and streamtools introduces dynamics that would confuse your analysis. For example, it's quite common for a system to wait until it has a collection of messages from its sensor before it makes an HTTP request to post those messages to the next stage. This means that by the time those messages make it to a streamtools pattern, they're artifcially grouped together into little pulses. You can use the sync block to recover the original stream generated by the sensor. 
    * Rules:
        * `Path`: [gojee](https://github.com/nytlabs/gojee) path. This must point at a UNIX epoch time in milliseconds
        * `Lag`: duration string

* **set**. This stores a [set](http://en.wikipedia.org/wiki/Set_(mathematics\)) of values as specified by the block's `Path`. Add new members through the (idempotent) ADD route. If you send a message through the ISMEMBER route, the block will emit true or false. You can also query the cardinality of the set. 
    * Rules:
        * `Path`: [gojee](https://github.com/nytlabs/gojee) path 

* **cache**. Stores string values against keys. Send a key to the `lookup` route and the value against that key will be emitted.
    * Rules:
        * `KeyPath`: [gojee](https://github.com/nytlabs/gojee) path to the element of the inbound message to use as key
        * `ValuePath`: [gojee](https://github.com/nytlabs/gojee) path to the element to store in the cache

* **queue**. This block represents a FIFO queue. You can push new messages onto the queue via the PUSH in route. You can pop messages off the queue either by hitting the POP inbound route, causing the block to emit the next message on its OUT route, or you can make a GET request to the POP query route and the block will respond with the next message. You can also peek at the next message using the PEEK query route. 

* **tolog**. Send messages to the log. This is a quick way to look at the data in your stream.

#### Pack

The three pack blocks group messages together in different ways. They operate similarly to an online "group-by" operation, but care needs to be taken in the stream setting as we have to decide when to emit the "packed" message. 

* **packbycount**. Groups messages into an array, emitting collected messages once specified MaxCount is reached.
    * Rules:
        * `MaxCount`: number of messages to group and emit at a time

* **packbyinterval**. Groups messages into an array, emitting collected messages at each specified interval.
    * Rules:
        * `Interval`: duration string (`1s`)
        
* **packbyvalue**. Groups messages with common value for a given key. Once we haven't seen any messages with that value for the given duration, it emits the collection.
    * Rules:
        * `EmitAfter`: duration string (`1s`)
        * `Path`: [gojee](https://github.com/nytlabs/gojee) path

     Our main use case for this at the NYT is to create per-reader reading sessions. So we set the `Path` to our user-id and we emit after 20 minutes of not hearing anything from that reader's user-id. Every page-view our readers generate get packed into a per-reader message, generating a stream of reading sessions. 


### Data Stores

These blocks send and retrieve data from various data stores.

* **toElasticsearch**. Send JSON to an [elasticsearch](http://www.elasticsearch.org/) instance.
    * Rules:
        * `Index`: 
        * `Host`: 
        * `IndexType`: 
        * `Port`: 

* **toFile**. Writes a message as JSON to a file. Each message becomes a new line of JSON. 
    * Rules:
        * `Filename`: file to write to

* **toMongoDB**. Saves messages to a [MongoDB](https://www.mongodb.org/) instance or a cluster. The messages can be saved as they come or in bulk depending on the user's needs.
    * Rules:
        * `Host`: the host string for the an instance, e.g. ```localhost:27107```, or a replicaset or a cluster, e.g. ```mongohost1.example.com:27017```, ```mongohost2.example.com:27017```, ```mongoarbiter1.example.com```
        * `Database`: database to which the documents should be written to.
        * `Collection`: collection to which the documents should be written to under the specified database.
        * `BatchSize`: the number of documents to be written together at any time in bulk. If the value is set to <= 1, the documents will be written one at a time. 

* **redis**. Sends arbitrary commands to redis. You can add or retrieve data from redis with this block.
    * Rules:
        * `Server`: The host string including port, defaults to ```localhost:6379```  
        * `Command`: Just the command, without arguments, to send to redis. Examples below.
        * `Arguments`: (optional) Array of options to send along with the command.
        * `Password`: (optional) Specify if your redis instance requires a password to connect.

```
{
  Arguments: [
    "'foo'",
    "'bar'",
    "'baz'"
  ],
  Command: "SADD",
  Password: "",
  Server: "localhost:6379"
}
```

### Network I/O

* **webRequest**. This blocks aspires to be curl inside streamtools. You can use the webRequest block to make custom requests to either a specific URL, or to a URL found in incoming messages in streamtools. You can also specify custom headers and scope the body of incoming messages for POST and PUT requests.
    * Rules:
    	* Use either Url **or** UrlPath. You can't use both :)
          * `Url`: a fully formed URL. 
          * `UrlPath`: [gojee](https://github.com/nytlabs/gojee) path to a fully formed URL found in an incoming message.
		* `Headers`: any http headers you wish to send in the request, represented in JSON. Example below.
		* `Method`: defaults to GET, select from a list that includes commonly used HTTP methods.
		* `BodyPath`: used only in POST and PUT requests, defaults to `.` (the entire incoming message), this data is sent with the request as the request body.

```
{
	"BodyPath": ".",
	"Headers": {
		"Content-Type": "application/json",
		"User-Agent": "I am not a robot"
	},
	"Method": "GET",
	"Url": "http://localhost:7070/library",
	"UrlPath": ""
}
```

* **fromEmail**. This block connects to the given IMAP server with the given credentials. Once connected, it idles on that connection and emits any unread emails into streamtools. Once messages have been pulled, the block marks them as read. All email messages will contain the `from`, `to`, `subject`, `internal_date` and `body` fields.
    * Rules:
        * `Host`: hostname of the IMAP server. Defaults to Gmail (imap.gmail.com)
        * `Username`: user for the email account.
        * `Password`: password for the email account.
        * `Mailbox`: the mailbox to pull email from. Defaults to 'INBOX' which is the main mailbox for Gmail.

* **fromUDP**. Listens for messages sent over UDP. Each message is emitted into streamtools.
    * Rules:
        * `ConnectionString`: host and port to connect to. Example: 127.0.0.1:0

* **fromHTTPStream**. This block allows you to listen to a long-lived http stream. Each new JSON that appears on the stream is emitted into streamtools. Try using the 1.usa.gov endpoint, available at ` http://developer.usa.gov/1usagov`. 
    * Rules:
        * `Endpoint`: endpoint string
        * `Auth`: authorisation string

* **fromPost**. This block emits any message that is POSTed to its IN route. This block isn't strictly needed as you can POST JSON to any inbound route on any block. Having said that, sometimes it's a bit clearer to have a dedicated block that listens for data.

* **fromWebsocket**. Connects to an existing websocket. Each message it hears from the websocket is emitted into streamtools. 
    * Rules:
        * `url`: address of the websocket.

* **fromHTTPGetRequest**. This block, when a GET request is made to the block's QUERY endpoint, emits that request into streamtools. The request can be handled by the ```toHTTPGetRequest``` block. 

* **toHTTPGetRequest**. This block responds to an HTTP GET request that has been generated by ```fromHTTPGetRequest```. The inbound message needs to contain both the original request and the message you want to respond with.
    * Rules:
        * `RespPath`: path to the HTTP request.
        * `MsgPath`: path to the message you want to respond with on the HTTP request.

* **getHTTP**. The getHTTP block makes an HTTP GET request to a URL you specify in the inbound message. It is necessary for the HTTP endpoint to serve JSON. This block forms the backbone for any sort of polling pattern.
    * Rules:
        * `Path`: [gojee](https://github.com/nytlabs/gojee) path to a fully formed URL. 

### Parsers

These blocks turn icky data into lovely json.

* **parsecsv**
* **parsexml**


### Queues

These blocks read from or write into external queue systems.

* **fromnsq**. This block implements an NSQ reader. For more details on how to specify the rule for this block check out the [NSQ docs](http://bitly.github.io/nsq/). 
    * Rules:
        * `ReadTopic`: topic to read from.
        * `LookupdAddr`: nsqlookupd addresss
        * `ReadChannel`: name of the channel 
        * `MaxInFlight`: how many messages to take from the queue at a time. (`0`)

* **toNSQ**. Send messages to an existing [NSQ](http://bitly.github.io/nsq/) system.
    * Rules:
        * `Topic`: topic you will write to
        * `NsqdTCPAddrs`: address of the NSQ daemon.
        
* **toNSQMulti**. Send messages to an NSQ system in batches. This is useful if you have a fast (>1KHz) stream of data you need to send to NSQ. This block gathers messages for `Interval` time and then sends. It emits immediately if the block gets more than `MaxBatch` messages.
    * Rules:
        * `Topic`: topic you will write to
        * `Interval`: duration string (`1s`)
        * `NsqdTCPAddrs`: address of the NSQ daemon.
        * `MaxBatch`: size of largest batch (`100`)

* **toBeanstalkd**. Send jobs to an existing [beanstalkd](https://github.com/kr/beanstalkd/) server.
    * Rules:
        * `Host`: the Host and port of the beanstalkd server e.g. ```127.0.0.1:11300```
        * `TTR`: Time to Run. is an integer number of seconds to allow a worker to run this job. This time is counted from the moment a worker reserves a job. If the worker does not delete, release, or bury the job within <TTR> seconds, the job will time out and the server will release the job.
        * `Tube` : beanstalkd tube to send jobs to. if left blank, jobs are sent to the default tube.
        
* **fromSQS**. This block connects to an [Amazon Simple Queueing System](http://aws.amazon.com/sqs/) queue. Messages from SQS are XML; this block extracts the message string from this XML, which it assumes is newline separated JSON. Each JSON is emitted into streamtools as a separate message. See the [SQS docs](http://aws.amazon.com/documentation/sqs/) for more information about the rules of this block.
    * Rules:
        * `SignatureVersion`: the version number of the signature hash Amazon is expecting for this queue (`4`)
        * `AccessKey`: your access key
        * `MaxNumberOfMessages`: how many messages to pull off the queue at a time (`10`)
        * `APIVersion`: what version of the API are you using (`2012-11-05`)
        * `SQSEndpoint`: the endpoint (ARM) of the SQS queue you are reading
        * `WaitTimeSeconds`: how long to wait between polling (`0`)
        * `AccessSecret`: your access secret

### Stats

* **count**. This block counts the number of messages it has seen over the specified `Window`. 
    * Rules:
        * `Window`: duration string (`0`)

* **histogram**. Build a non-staionary histogram of the inbound messages. Currently this only works with discrete values.
    * Rules:
        * `Path`: [gojee](https://github.com/nytlabs/gojee) path to the value over which you'd like to build a histogram.
        * `Window`: duration string specifying how long to retain messages in the histogram (`0`)

* **timeseries**. This block stores an array of the value specified by `Path` along with the timestamp at the time the message arrived.
    * Rules:
        * `Path`: [gojee](https://github.com/nytlabs/gojee) path
        * `NumSamples`: how many samples to store (`0`)

* **kullbackleibler**. Calculates the [Kullback Leibler divergence](http://en.wikipedia.org/wiki/Kullback%E2%80%93Leibler_divergence) between two distributions p and q. The two distributions must mimic the output from the ```histogram``` block.
    * Rules:
        * `QPath`: [gojee](https://github.com/nytlabs/gojee) path to the q distribution. 
        * `PPath`: [gojee](https://github.com/nytlabs/gojee) path to the p distribution. 

* **movingaverage**. Performs a [moving average](http://en.wikipedia.org/wiki/Moving_average) of the values specified by the `Path` over the duration of the `Window`.
    * Rules:
        * `Path`: [gojee](https://github.com/nytlabs/gojee) path
        * `Window`: duration string
    
* **zipf**. This block draws a random number from a [Zipf-Mandelbrot](http://en.wikipedia.org/wiki/Zipf%E2%80%93Mandelbrot_law) distribution when polled.
    * Rules:
        * `s`: (`2`)
        * `v`: (`5`)
        * `N`: (`99`)

* **gaussian**. This block draws a random number from the [Gaussian](http://en.wikipedia.org/wiki/Gaussian_distribution) distribution when polled.
    * Rules:
        * `StdDev`: (`1`)
        * `Mean`: (`0`)

* **poisson**. This block draws a random number from a [Poisson](http://en.wikipedia.org/wiki/Poisson_distribution) distribution when polled.
    * Rules:
        * `Rate`: (`1`)

* **categorical**. This block draws a random number from a [Categorical](http://en.wikipedia.org/wiki/Categorical_distribution) distribution when polled.
    * Rules:
        * `Weights`: (`[1]`) - a list of weighting parameters. The number drawn from this distribution corresponds to the index of this list. These weights are automatically normalised to sum to one. 
        
## Interface

Streamtool's GUI aims to be responsive and informative, meaning that you can both create and interrogate a live streaming system. At the same time, it aims to be as minimal as possible - the GUI posses a very tight relationship with the underlying streamtools architecture, enabling users of streamtools to see and understand the execution of the system.

### Blocks

To make a block, double click anywhere on the background. Type the name of the block you'd like and press enter.

![create](https://f.cloud.github.com/assets/597897/2443728/161a3708-ae3e-11e3-9ac7-7e3062720bd7.gif)

To delete a block you don't like anymore, click on it and press the delete (backspace) button on your keyboard.

![delete](https://f.cloud.github.com/assets/597897/2444176/d2a457ca-ae4b-11e3-81a0-beb016f58bb6.gif)

To move a block around, simply drag it about the place.

![drag_2](https://f.cloud.github.com/assets/597897/2443683/7d8d0cbe-ae3c-11e3-91ad-852e830dbeb8.gif)

### Connections

To connect two blocks together, first click on an outbound route on the *bottom* of the block you want to connect from. Almost always this route will be labelled `OUT` when you mouse over it. Then click on an inbound route on the *top* of another block. There can be a few inbound routes; common ones are `IN`, `RULE`, and `POLL`. This will create a connection between the blocks.

![connect_2](https://f.cloud.github.com/assets/597897/2443787/8070125c-ae3f-11e3-92ba-8a69f3ef24dc.gif)

### Rules

To set a block's rules, double click it. This will open a window where you can enter rules. When you're done entering rules, hit the update button.

![update_rule](https://f.cloud.github.com/assets/597897/2443884/f2fa3e62-ae42-11e3-8b25-486e8f034677.gif)

### Queries

You can query a block's rules, or any other queryable route a block has, by clicking on the red squares on the right of the block. These will open a window that shows a JSON representation of that information. An example of a queryable route is `COUNT` for the count block. If you click on the little red square associated with the `COUNT` route, then you'll get a JSON representation of that block's current count.

![query](https://f.cloud.github.com/assets/597897/2444136/d4178984-ae4a-11e3-8861-869e6dd472b3.gif)

### Messages

To see the last message that passed through a connection, click and drag the connection's rate estimate. This creates a window containing the JSON representation of the last message to pass through that connection.

![last_message](https://f.cloud.github.com/assets/597897/2443597/b1c2411e-ae39-11e3-9fb4-429e39548620.gif)


## API

Streamtools provides a full RESTful HTTP API allowing the developer to programatically control all aspects of streamtools. The API can be broken up into three parts: those endpoints that general aspects of streamtools, those that control blocks, and those that control connections.

If you are running streamtools locally, using the default port, all of the GET endpoints can be queried either by visiting in a browser:

    http://localhost:7070/{endpoint}

For example, if you wanted to see the streamtools library, visit `http://localhost:7070/library`.

The POST endpoints are expecting you to send data. To use these you'll need to use the command line and a program called `curl`. For example, to create a new `tofile` block you need to send along the JSON definition of the block, like this:

    curl http://localhost:7070/blocks -d'{"Type":"tofile","Rule":{"Filename":"test.json"}}'

This POSTs the JSON `{"Type":"tofile","Rule":{"Filename":"test.json"}}` to the `/blocks` endpoint.


### General

GET `/library`

The library endpoint returns a description of all the blocks available in the version of streamtools that is runnning.

GET `/version`

The version endpoint returns the current version of streamtools.

GET `/export`

Export returns a JSON representation of the current streamtools pattern.

POST `/import`

Import accepts a JSON representation of a pattern, creating it in the running streamtools instance. Any block ID collissions are resolved automatically, meaning you can repeatedly import the same pattern if it's useful.

### Blocks

A block's JSON representation uses the following schema:

```
{
  "Id":
  "Type":
  "Rule":{ ... }
  "Position":{
    "X":
    "Y":
  }
}
```

Only `Type` is required, everything will be automatically generated if you don't specify them. The `Id` is used to uniquely identify that block within streamtools. This is normally just a number but can be any string. `Type` is the type of the block, selected from the streamtools library. `Rule` specifies the block's rule, which will be different for each block. Finally `Position` specifies the x and y coordinates of the block from the top left corner of the screen.

* POST `/blocks`
	* To create a new block, simply POST its JSON representation as described above to the `/blocks` endpoint.
* GET `/blocks/{id}`
	* Returns a JSON representation of the block specified by `{id}`.
* DELETE `/blocks/{id}`
	* Deletes the block specified by `{id}`.
* POST `/blocks/{id}/{route}`
	* Send data to a block. Each block has a set of default routes ("in","rule") and optional routes ("poll"), as well as custom rotues that defined by the block designer as they see fit. This will POST your JSON to the block specified by `{id}` via route `{route}`.
* GET `/blocks/{id}/{route}`
	* Recieve data from a block. Use this endpoint to query block routes that return data. The only default route is `rule` which, in response to a GET query, will return the block's current rule.

### Connections

A connection's JSON representation uses the following schema:
```
{
  Id:
  FromId:
  ToId:
  ToRoute:
}
```
Here, only `Id` is optional. `Id` is used to uniquely refer to the connection inside streamtools. `FromId` refers to the block that data is flowing from. `ToId` refers to the block the data is flowing to. `ToRoute` tells the connection which inbound route to send data to.

* POST `/connections`
	* Post a connection's JSON representation to this endpoint to create it.
* GET `/connections`
	* Lists all the current connections.
* GET `/connections/{id}`
	* Returns the JSON representation of the connection specified by `{id}`.
* DELETE `/connections/{id}`
	* Deletes the connection specified by `{id}`.
* GET `/connections/{id}/{route}`
	* Query a connection via its routes. Each connection has a `rate` route which will return an estimate of the rate of messages coming through it and a `last` route which will return the last message it saw.

### Messages

Every block that as an `OUT` route also has a websocket and a long-lived HTTP connection associated with it. These are super useful for getting data out of streamtools.

WEBSOCKET `/ws/{id}`

a websocket emitting every message sent on the block's `OUT` route.

GET `/stream/{id}`

a long-lived HTTP stream of every message sent on the block's `OUT` route.

## Command Line

The streamtools server is completely contained in a single binary called `st`. It has a number of options:

* `--port=7070` - specify a port number to run on. Default is 7070.
* `--domain=localhost` - if you're accessing streamtools through a URL that's not `localhost`, you need to specify it using this option.


## More Info

For more info see [Introducing Streamtools](http://blog.nytlabs.com/2014/03/12/streamtools-a-graphical-tool-for-working-with-streams-of-data/) on The New York Times R&D Labs blog.

For background on responsive programming tools see Bret Victor's [learnable programming](http://worrydream.com/#!/LearnableProgramming). 

If you're interested in learning more about visual programming languages, check out Interface Vision's [fantastic roundup dating back to 1963](http://blog.interfacevision.com/design/design-visual-progarmming-languages-snapshots/). 
