package pubsub

import (
	"fmt"

	"github.com/go-redis/redis/v7"
)

// RedisService is a service that handles our messages in the channel
type RedisService struct {
	Client *redis.Client
	Addr   string
}

// Connect instantiates a new client to the Redis database
// Returns nil if the connection was successful and an error if not
func (rs *RedisService) Connect() error {
	rs.Client = redis.NewClient(&redis.Options{
		Addr:     rs.Addr,
		Password: "",
		DB:       0,
	})
	return rs.TestConnection()
}

// TestConnection checks if you have a valid Redis client
// Returns an error if something is wrong
func (rs *RedisService) TestConnection() error {
	if rs.Client == nil {
		return fmt.Errorf("Your Redis client is nil, have you called RedisService.Connect()?")
	}
	reply, err := rs.Client.Ping().Result()
	if err == nil && reply == "PONG" {
		return nil
	}
	return err
}

// Subscribe subscribes to the given Redis channel and redirects messages
// to the given Go channel in a go routine
func (rs *RedisService) Subscribe(rsCh string, goCh chan string) error {
	// Make sure a valid client is present
	err := rs.TestConnection()
	if err != nil {
		return err
	}

	// Listen for messages in a go routine and pass them to the go channel
	go func() {
		sub := rs.Client.Subscribe(rsCh)
		msg, _ := sub.ReceiveMessage()
		switch msg.Channel {
		case "sophie":
			goCh <- msg.Payload
		}
	}()
	return nil
}
