package library

import (
	"encoding/json"
	"errors"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"github.com/nytlabs/streamtools/st/util"
)

// specify those channels we're going to use to communicate with streamtools
type ToMongoDB struct {
	blocks.Block
	queryrule chan chan interface{}
	inrule    chan interface{}
	in        chan interface{}
	quit      chan interface{}
}

// we need to build a simple factory so that streamtools can make new blocks of this kind
func NewToMongoDB() blocks.BlockInterface {
	return &ToMongoDB{}
}

// Setup is called once before running the block. We build up the channels and specify what kind of block this is.
func (b *ToMongoDB) Setup() {
	b.Kind = "ToMongoDB"
	b.in = b.InRoute("in")
	b.inrule = b.InRoute("rule")
	b.queryrule = b.QueryRoute("rule")
	b.quit = b.Quit()
}

// Run is the block's main loop. Here we listen on the different channels we set up.
func (b *ToMongoDB) Run() {
	var collectionname string
	var dbname string
	var collection *mgo.Collection
	var session *mgo.Session
	var host = ""
	var err error
	for {
		select {
		case msgI := <-b.inrule:
			// set host string for MongoDB server
			host, err = util.ParseString(msgI, "Host")
			if err != nil {
				b.Error(err.Error())
				continue
			}
			// set collection name
			dbname, err = util.ParseString(msgI, "Database")
			if err != nil {
				b.Error(err.Error())
				continue
			}
			// set time to reserve
			collectionname, err = util.ParseString(msgI, "Collection")
			if err != nil {
				b.Error(err.Error())
			}
			// create MongoDB connection
			session, err = mgo.Dial(host)
			if err != nil {
				// swallowing a panic from mgo here - streamtools must not die
				b.Error(errors.New("Could not initiate connection with MongoDB service"))
				continue
			}
			// use the specified DB and collection
			collection = session.DB(dbname).C(collectionname)
		case <-b.quit:
			// close connection to MongoDB and quit
			if session != nil {
				session.Close()
			}
			return
		case msg := <-b.in:
			// deal with inbound data
			msgStr, err := json.Marshal(msg)
			if err != nil {
				b.Error(err)
				continue
			}
			if session != nil {

				err := collection.Insert(bson.M{"val":string(msgStr)})
				if err != nil {
					b.Error(err.Error())
				}
			} else {
				b.Error(errors.New("MongoDB connection not initated or lost. Please check your MongoDB server or block settings."))
			}
		case respChan := <-b.queryrule:
			// deal with a query request
			respChan <- map[string]interface{}{
				"Host": host,
				"Database": dbname,
				"Collection":  collectionname,
			}
		}
	}
}
