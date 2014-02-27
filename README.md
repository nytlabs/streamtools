# streamtools

[![Build Status](https://travis-ci.org/nytlabs/streamtools.png?branch=master)](https://travis-ci.org/nytlabs/streamtools)

Streamtools is a graphical toolkit for dealing with streams of data. Streamtools makes it easy to explore, analyse, modify and learn from streams of data.

## Blocks



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
