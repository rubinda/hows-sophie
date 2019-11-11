package sophie

const (
	// ProtocolICMP is the code for the ICMP packet
	ProtocolICMP = 1
	// StatusFile holds the status messages
	StatusFile = "configs/statuses.yml"
	// ConfigFile holds some general parameters
	ConfigFile = "configs/config.yml"
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

// TweetService provides an interface to a service that posts tweets.
// Todo change the return type to a more general object
type TweetService interface {
	PostTweet(msg string) error
}

// MessageService provides an interface for implementing a message
// broker, to handle PC events.
type MessageService interface {
	Connect() error
	Subscribe(rsCh string) error
}
