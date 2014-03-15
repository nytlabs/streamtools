# streamtools

[![Build Status](https://travis-ci.org/nytlabs/streamtools.png?branch=master)](https://travis-ci.org/nytlabs/streamtools)

Streamtools is a graphical toolkit for dealing with streams of data. Streamtools makes it easy to explore, analyse, modify and learn from streams of data.

Quick links from our wiki:

* [how to use the GUI](https://github.com/nytlabs/streamtools/wiki/GUI)
* [block documentation](https://github.com/nytlabs/streamtools/wiki/blocks)
* [how to use the API](https://github.com/nytlabs/streamtools/wiki/API)
* [how to compile](https://github.com/nytlabs/streamtools/wiki/how-to-compile)


## Getting Started - the nuts and bolts

### quick start

* download `st` from the [streamtools releases](https://github.com/nytlabs/streamtools/releases) page
* run `st` locally or on server
* in a browser, visit port 7070 of the machine you ran `st` on.

### longer description

Mostly, you'll interact with streamtools in the browser. A server program, called `st` runs on a computer somewhere that serves up the streamtools webpage. Either it will be on your local machine, or you can put it on a remote machine somewhere - we often run it on a virtual computer in Amazon's cloud so we can leave streamtools running for long periods of time. To begin with, though, we'll assume that you're running streamtools locally, on a machine you can touch. We're also going to assume you're running OSX or Linux - if you're a Windows user you will need to compile the code yourself.

So, first of all, you need to download the streamtools server. It's just a single file, and you can find the latest release on [github](https://github.com/nytlabs/streamtools/releases). Download this file, and move it to your home directory. Now, open a terminal and run the streamtools server by typing `~/st`. You should see streamtools start up, telling you it's running on port 7070.

Now, open a browser window and point it at [localhost:7070](http://localhost:7070/). You should see a (nearly) blank page. At the bottom you should see a status bar that says `client: connected to Streamtools` followed by a version number. Congratulations! You're in.

## Command Line Options

The streamtools server is completely contained in a single binary called `st`. It has a number of options:

* `--port=7070` - specify a port number to run on. Default is 7070.
* `--domain=localhost` - if you're accessing streamtools through a URL that's not `localhost`, you need to specify it using this option.

## How Streamtools works

Streamtools' basic paradigm is straightforward: data flows from *blocks* through *connections* to other blocks. A block perfoms some operation on each message it recieves, and that operation is defined by the block's *type*. Each block has zero or more *rules* which define that block's behaviour. Each block has a set of named *routes* that can recieve data, emit data, or respond to queries.

A block's rule can be set directly by double clicking on a block and typing in the rule manually. Alternatively, a block's rule can be set by sending an appropriately formed message to the block's `rule` route.

You can connect blocks together, via their routes, using connections. You can connect to any inbound route, and so data flowing through streamtools can be used to set the rules of the blocks in the running pattern.

We call a collection of connected blocks a *pattern*, and it is possible to export and import whole patterns from a running instance of streamtools. Together, these 5 concepts: blocks, rules, connections, routes and patterns form the basic vocabulary we use to talk about streamtools.

# References

* For background on responsive programming tools see Bret Victor's [learnable programming](http://worrydream.com/#!/LearnableProgramming).
