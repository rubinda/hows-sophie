package pubsub

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"

	"github.com/go-redis/redis/v7"
	sophie "github.com/rubinda/hows-sophie"
	"gopkg.in/yaml.v2"
)

// RedisService is a service that handles our messages in the channel
type RedisService struct {
	Client     *redis.Client
	Addr       string
	StatusMsgs *sophie.Status
}

// Connect instantiates a new client to the Redis database and
// reads the statuses into the service.
// Returns nil if the connection was successful and an error if not
func (rs *RedisService) Connect() error {
	rs.Client = redis.NewClient(&redis.Options{
		Addr:     rs.Addr,
		Password: "",
		DB:       0,
	})

	// Read the YAML statuses from the file and read it into the struct
	rs.StatusMsgs = &sophie.Status{}
	rawYaml, err := ioutil.ReadFile(sophie.StatusFile)
	if err != nil {
		// Todo better error handling
		panic(err)
	}
	err = yaml.Unmarshal(rawYaml, rs.StatusMsgs)
	if err != nil {
		// Todo better error handling
		panic(err)
	}
	return rs.testConnection()
}

// TestConnection checks if you have a valid Redis client
// Returns an error if something is wrong
func (rs *RedisService) testConnection() error {
	if rs.Client == nil {
		return fmt.Errorf("Your Redis client is nil, have you called RedisService.Connect()?")
	}
	reply, err := rs.Client.Ping().Result()
	if err == nil && reply == "PONG" {
		return nil
	}
	return err
}

// PickMessage selects a random message from the status file based
// on the status received.
//		<serviceType>:<eventName>
// e.g. 'status:offline' will pick a message about Sophie shutting down
func (rs *RedisService) pickTweet(serviceType string, eventName string) string {
	// Todo: remove hardcoded paths
	var statusMsg string
	switch serviceType {
	case "status":
		switch eventName {
		case "online":
			statusMsg = rs.StatusMsgs.Online.Normal[rand.Intn(len(rs.StatusMsgs.Online.Normal))]
		case "offline":
			statusMsg = rs.StatusMsgs.Offline[rand.Intn(len(rs.StatusMsgs.Offline))]
		}
	}
	return statusMsg
}

// Subscribe subscribes to the given Redis channel and redirects messages
// to the given Go channel in a go routine
func (rs *RedisService) Subscribe(rsCh string, ts sophie.TweetService) error {
	// Make sure a valid client is present
	err := rs.testConnection()
	if err != nil {
		return err
	}

	// Listen for messages in a go routine and pass them to the go channel
	go func() {
		colon := regexp.MustCompile(`:`)
		sub := rs.Client.Subscribe(rsCh)
		for {
			msg, _ := sub.ReceiveMessage()
			switch msg.Channel {
			case "sophie":
				// Decode the message, prepare the appropriate status message
				// and post a tweet. Events are always given in the following
				// format:
				// <service_type>:<event_name>
				// e.g. status:online -> Sophie has just powered on
				ep := colon.Split(msg.Payload, 2)
				statusMsg := rs.pickTweet(ep[0], ep[1])
				err = ts.PostTweet(statusMsg)
				if err != nil {
					fmt.Printf("Problem occured while tweeting: %v\n", err)
				}
			}
		}
	}()
	return nil
}
