package tweet

import (
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	sophie "github.com/rubinda/hows-sophie"
)

// TwitterService implements the interface for posting tweets
type TwitterService struct {
	Client *twitter.Client
}

// SetTwitterClient creates a new Twitter client for the given credentials
func (ts *TwitterService) SetTwitterClient(creds *sophie.Credentials) {
	config := oauth1.NewConfig(creds.ConsumerKey, creds.ConsumerSecret)
	token := oauth1.NewToken(creds.AccessToken, creds.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	ts.Client = twitter.NewClient(httpClient)
}

// PostTweet adds a new message to the twitter account
func (ts *TwitterService) PostTweet(msg string) error {
	if ts.Client == nil {
		return fmt.Errorf("Twitter client is nil, have you called SetTwitterClient?")
	}

	_, resp, err := ts.Client.Statuses.Update("Success, my bot can create tweets!", nil)
	if err != nil {
		fmt.Printf("Error occured: %v", resp)
		return err
	}
	return nil
}
