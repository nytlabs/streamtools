# streamtools

[![Build Status](https://travis-ci.org/nytlabs/streamtools.png?branch=master)](https://travis-ci.org/nytlabs/streamtools)

Streamtools is a graphical toolkit for dealing with streams of data. Streamtools makes it easy to explore, analyse, modify and learn from streams of data.

## How Streamtools works

Streamtools' basic paradigm is straightforward: data flows from *blocks* through *connections* to other blocks. A block perfoms some operation on each message it recieves, and that operation is defined by the block's *type*. Each block has zero or more *rules* which define that block's behaviour. Each block has a set of named *routes* that can recieve data, emit data, or respond to queries.

A block's rule can be set directly by double clicking on a block and typing in the rule manually. Alternatively, a block's rule can be set by sending an appropriately formed message to the block's `rule` route.

You can connect blocks together, via their routes, using connections. You can connect to any inbound route, and so data flowing through streamtools can be used to set the rules of the blocks in the running pattern.

We call a collection of connected blocks a *pattern*, and it is possible to export and import whole patterns from a running instance of streamtools. Together, these 5 concepts: blocks, rules, connections, routes and patterns form the basic vocabulary we use to talk about streamtools.

## Getting Started - the nuts and bolts

Mostly, you'll interact with streamtools in the browser. A server program, called `st` runs on a computer somewhere. Either it will be on your local machine, or you can put it on a remote machine somewhere - we often run it on a virtual computer in Amazon's cloud so we can leave streamtools running for long periods of time. To begin with, though, we'll assume that you're running streamtools locally, on a machine you can touch. We're also going to assume you're running OSX or Linux - if you're a Windows user you will need to compile the code yourself.

So, first of all, you need to download the streamtools server. It's just a single file, and you can find the latest release on [github](https://github.com/nytlabs/streamtools/releases). Download this file, and move it to your home directory. Now, open a terminal and run the streamtools server by typing `~/st`. You should see streamtools start up, telling you it's running on port 7070.

Now, open a browser window and point it at [localhost:7070](http://localhost:7070/). You should see a (nearly) blank page. At the bottom you should see a status bar that says `client: connected to Streamtools` followed by a version number. Congratulations! You're in.

## GUI

Streamtool's GUI aims to be responsive and informative, meaning that you can both create and interrogate a live streaming system. At the same time, it aims to be as minimal as possible - the GUI posses a very tight relationship with the underlying streamtools architecture enabling users of streamtools to see and understand the execution of the system.

### make a block

To make a block, double click anywhere on the background. Type the name of the block you'd like and press enter.

### connect two blocks

To connect two blocks together, first click on an outbound route on the *bottom* of the block you want to connect from. Almost always this route will be labelled `OUT` when you mouse over it. Then click on an inbound route on the *top* of another block. There can be a few inbound routes; common ones are `IN`, `RULE`, and `POLL`. This will create a connection between the blocks.

### set the rule of a block

To set a block's rules, double click it. This will open a window where you can enter rules. When you're done entering rules, hit the update button.

### query a block

You can query a block's rules, or any other queryable route a block has, by clicking on the red squares on the right of the block. These will open a window that shows a JSON representation of that information. An example of a queryable route is `COUNT` for the count block. If you click on the little red square associated with the `COUNT` route, then you'll get a JSON representation of that block's current count.

### delete a block

To delete a block you don't like anymore, click on it and press the delete (backspace) button on your keyboard.

### move a block

To move a block around, simply drag it about the place.

### see the last message that passed through a connection

To see the last message that passed through a connection, click and drag the connection's rate estiamte. This creates a window containing the JSON representation of the last message to pass through that connection.  

## Blocks

Streamtools contains an ever-increasing set of individual blocks you can use to build patterns. Each block normally has one or more of the following routes, though blocks can also have custom routes as necessary.

* `in` : a block accepts data from this route
* `out` : a block will emit data on this route
* `rule` : a block will update its rule using message from this route
* `poll` : a block will emit its state through its `out` route when it recieves any signal on this route

We break blocks up roughly into generator, source, sink, transfer and state blocks. While there is nothing inherent to streamtools that enforces these ideas, we've found that blocks naturally tend to these categories.

### Generator Blocks

* ticker

### Source Blocks

* fromSQS
* fromNSQ
* fromHTTPStream

### Sink Blocks

* toNSQ
* toFile
* toLog

### Transfer Blocks

* map
* filter
* sync
* getHTTP
* gaussian
* zipf
* poisson

### State Blocks

* count
* histogram

## Patterns

### Counting

One of the most basic patterns when dealing with stream

![count](https://raw.github.com/mikedewar/streamtools/redesign/examples/count.png)

[counting example](https://raw.github.com/mikedewar/streamtools/redesign/examples/count.json)

## Command Line Options

The streamtools server is completely contained in a single binary called `st`. It has a number of options:

* *--port=7070* - specify a port number to run on. Default is 7070.
* *--domain=localhost* - if you're accessing streamtools through a URL that's not `localhost`, you need to specify it using this option.

## API

Streamtools provides a full RESTful HTTP API allowing the developer to programatically control all aspects of streamtools. The API can be broken up into three parts: those endpoints that general aspects of streamtools, those that control blocks and those that control connections.

If you are running streamtools locally, using the default port, all of the GET endpoints can be queried either by visiting in a browser:

    http://localhost:7070/{endpoint}

For example, if you wanted to see the streamtools library, visit `http://localhost:7070/library`.

The POST endpoints are expecting you to send data. To use these you'll need to use the command line and a program called `curl`. For example, to create a new `tofile` block you need to send along the JSON definition of the block, like this:

    curl http://localhost:7070/blocks -d'{"Type":"tofile","Rule":{"Filename":"test.json"}}'

This POSTs the JSON `{"Type":"tofile","Rule":{"Filename":"test.json"}}` to the `/blocks` endpoint.


### streamtools

GET `/library`

The library endpoint returns a description of all the blocks available in the version of streamtools that is runnning.

GET `/version`

The version endpoint returns the current version of streamtools.

GET `/export`

Export returns a JSON representation of the current streamtools pattern.

POST `/import`

Import accepts a JSON representation of a pattern, creating it in the running streamtools instance. Any block ID collissions are resolved automatically, meaning you can repeatedly import the same pattern if it's useful.

### blocks

A block's JSON representation uses the following schema:

    {
      "Id":
      "Type":
      "Rule":{ ... }
      "Position":{
        "X":
        "Y":
      }
    }

Only `Type` is required, everything will be automatically generated if you don't specify them. The `Id` is used to uniquely identify that block within streamtools. This is normally just a number but can be any string. `Type` is the type of the block, selected from the streamtools library. `Rule` specifies the block's rule, which will be different for each block. Finally `Position` specifies the x and y coordinates of the block from the top left corner of the screen.

POST `/blocks`

To create a new block, simply POST its JSON representation as described above to the `/blocks` endpoint.

GET `/blocks/{id}`

Returns a JSON representation of the block specified by `{id}`.

DELETE `/blocks/{id}`

Deletes the block specified by `{id}`.

POST `/blocks/{id}/{route}`

Send data to a block. Each block has a set of default routes ("in","rule") and optional routes ("poll"), as well as custom rotues that defined by the block designer as they see fit. This will POST your JSON to the block specified by `{id}` via route `{route}`.

GET `/blocks/{id}/{route}`

Recieve data from a block. Use this endpoint to query block routes that return data. The only default route is `rule` which, in response to a GET query, will return the block's current rule.

### connections

A connection's JSON representation uses the following schema:

    {
      Id:
      FromId:
      ToId:
      ToRoute:
    }

Here, only `Id` is optional. `Id` is used to uniquely refer to the connection inside streamtools. `FromId` refers to the block that data is flowing from. `ToId` refers to the block the data is flowing to. `ToRoute` tells the connection which inbound route to send data to.

POST `/connections`

Post a connection's JSON representation to this endpoint to create it.

GET `/connections`

Lists all the current connections.

GET `/connections/{id}`

Returns the JSON representation of the connection specified by `{id}`.

DELETE `/connections/{id}`

Deletes the connection specified by `{id}`.

GET `/connections/{id}/{route}`

Query a connection via its routes. Each connection has a `rate` route which will return an estimate of the rate of messages coming through it and a `last` route which will return the last message it saw.

# References

* For background on responsive programming tools see Bret Victor's [learnable programming](http://worrydream.com/#!/LearnableProgramming).
