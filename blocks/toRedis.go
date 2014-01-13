package blocks

import (
  "log"
  "time"
  "github.com/garyburd/redigo/redis"
)

// Posts a message to Redis
func ToRedis(b *Block) {

	type toRedisRule struct {
		Server        string // defaults to "localhost:6379"
	}

	var rule *toRedisRule

	// TODO check the endpoint for happiness
	for {
		select {
		case msg := <-b.Routes["set_rule"]:
      if (rule == nil) {
        rule = &toRedisRule{"localhost:6379"}
      }
			unmarshal(msg, rule)

		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &toRedisRule{})
			} else {
				marshal(msg, rule)
			}
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

		  log.Println(msg)
      // Setup a pool to manage redis connections
      pool := &redis.Pool{
             MaxIdle: 3,
             IdleTimeout: 240 * time.Second,
             Dial: func () (redis.Conn, error) {
                 c, err := redis.Dial("tcp", rule.Server)
                 if err != nil {
                     return nil, err
                 }
                 return c, err
             },
             TestOnBorrow: func(c redis.Conn, t time.Time) error {
                 _, err := c.Do("PING")
                 return err
             },
         }
      log.Println(pool)

      // Get a connection from the pool
      conn := pool.Get()
      conn.Do("SET", "Key", "Value")
  	}
	}
}
