package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/rubinda/hows-sophie/pubsub"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"gopkg.in/yaml.v2"
)

const (
	// ProtocolICMP is the code for the ICMP packet
	ProtocolICMP = 1
	// ListenAddr is the default IPv4 address for listening
	ListenAddr = "0.0.0.0"
	// StatusFile holds the status messages
	statusFile = "statuses.yml"
)

// Credentials holds the tokens to access the Twitter API
type Credentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

// Status describes the object present in the 'statusFile'
// It holds the messages which are posted to Twitter and
// present in the previously mentined file.
type Status struct {
	Online struct {
		Late   []string `yaml:"late"`
		Normal []string `yaml:"normal"`
	} `yaml:"online"`
	Offline []string `yaml:"offline"`
}

// getTwitterClient returns a new Twitter client for the given credentials
func getTwitterClient(creds *Credentials) *twitter.Client {
	config := oauth1.NewConfig(creds.ConsumerKey, creds.ConsumerSecret)
	token := oauth1.NewToken(creds.AccessToken, creds.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	return twitter.NewClient(httpClient)
}

// Ping sends a simple echo request to the specified address and returns the
// address of the responder, the round trip time and an error if it occured.
//
// Sequence is used if more than 1 ping is transmitted in the same process`
// The timeout is how long we wait for the echo reply before returning a
// destination unreachable.
func ping(dest string, sequence int, timeoutSec int) (*net.IPAddr, time.Duration, error) {
	// Convert the int seconds to time.Duration
	timeout := time.Duration(timeoutSec)

	// Listen for potential ICMP replies
	conn, err := icmp.ListenPacket("ip4:icmp", ListenAddr)
	if err != nil {
		return nil, 0, err
	}
	defer conn.Close()

	// Check if the given address needs to be resolvd
	destAddr, err := net.ResolveIPAddr("ip4", dest)
	if err != nil {
		return nil, 0, err
	}

	// Compose a new ICMP Echo Request
	// Checksum is calculated automatically when the message is encoded
	// Data is empty for now
	echo := icmp.Message{
		Type:     ipv4.ICMPTypeEcho,
		Code:     0,
		Checksum: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid(),
			Seq:  sequence,
			Data: []byte(""),
		},
	}

	// Convert the message to bytes (also calculates checksum):
	// https://godoc.org/golang.org/x/net/icmp#Message.Marshal
	echoB, err := echo.Marshal(nil)
	if err != nil {
		return destAddr, 0, err
	}

	// Send the message to the destAddrination address
	start := time.Now()
	// TODO: check what the first parameter is (n?)
	_, err = conn.WriteTo(echoB, destAddr)
	if err != nil {
		return destAddr, 0, err
	}

	// Listen for a reply with a given timeout
	echoReplyB := make([]byte, 1500)
	err = conn.SetReadDeadline(time.Now().Add(timeout * time.Second))
	if err != nil {
		return destAddr, 0, err
	}
	// Read the reply and calculate the duration of the ping
	n, replier, err := conn.ReadFrom(echoReplyB)
	rtt := time.Since(start)
	if err != nil {
		if strings.Contains(err.(*net.OpError).Error(), "timeout") {
			return destAddr, rtt, fmt.Errorf("Destination host: %v unreachable (timeout %v)", destAddr, timeout*time.Second)
		}
		return destAddr, rtt, err
	}

	// Cast the Addr to IPAddr
	replierAddr := replier.(*net.IPAddr)

	// Parse the encoded message and check if it's a EchoReply
	echoReply, err := icmp.ParseMessage(ProtocolICMP, echoReplyB[:n])
	if err != nil {
		return replierAddr, rtt, err
	}
	switch echoReply.Type {
	case ipv4.ICMPTypeEchoReply:
		return replierAddr, rtt, nil
	default:
		return replierAddr, rtt, fmt.Errorf("Wrong response! \n\tWanted: %v \n\tGiven: %v",
			ipv4.ICMPTypeEchoReply, echoReply.Type)
	}
}

// isOnline checks wheter the specified host is reachable through the network
// (it sends some ping requests and checks for valid replies)
func isOnline(host string) (time.Duration, error) {
	// How long to wait before declaring unreachable state
	pingTimeout := 3
	// How many pings to send (isOnline if exists 1 that responded with reachable)
	pingCount := 4
	// The average rtt for the SUCCESSFUL pings
	rttAvg := time.Duration(0)
	// Count the successful pings
	goodPings := 0
	// The error when trying to ping and the round trip time for the ping
	var pingErr error
	var rtt time.Duration

	for seq := 1; seq <= pingCount; seq++ {
		_, rtt, pingErr = ping(host, seq, pingTimeout)
		if pingErr == nil {
			rttAvg += rtt
			goodPings++
		}
	}
	if goodPings > 0 {
		return rttAvg / time.Duration(goodPings), nil
	}
	return 0, pingErr
}

// tweetStatus posts a new tweet with the given message
func tweetStatus(tc *twitter.Client, msg string) (*twitter.Tweet, error) {
	tweet, resp, err := tc.Statuses.Update("just setting up my twttr", nil)
	if err != nil {
		fmt.Printf("Error occured: %v", resp)
		return nil, err
	}
	return tweet, nil
}

func main() {
	// Read the YAML statuses from the file and read it into the struct
	statuses := &Status{}
	rawYaml, err := ioutil.ReadFile(statusFile)
	if err != nil {
		// Todo better error handling
		panic(err)
	}
	err = yaml.Unmarshal(rawYaml, statuses)
	if err != nil {
		// Todo better error handling
		panic(err)
	}

	// Create a new client for the Redis DB
	// Todo move parameter to config file / ENV
	rdb := &pubsub.RedisService{
		Addr: "localhost:6379",
	}
	err = rdb.Connect()
	if err != nil {
		panic(err)
	}

	// Subscribe to the channel for Sophie
	// Todo move parameter to config file / ENV
	messages := make(chan string)
	rdb.Subscribe("sophie", messages)

	// Block the main thread so we can keep listening or messages
	forever := make(chan bool)

	// Whenever a message pops, handle it
	go func() {
		for d := range messages {
			fmt.Printf("Received message: %v", d)
		}
	}()

	fmt.Printf("Waiting for messages, to exit press CTRL+C")
	<-forever

	//scheduler := cron.New()
	// ToDo:
	// 		Function to check the current status of Sophie
	// 		Figure out a way to store the status of Sophie (REDIS!)
	// 		After X. PM (in the evening) check if Sophie is still up periodically until she is shutdown
	// 		Write some tests!
}
