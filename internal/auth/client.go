package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, fmt.Errorf("unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web %v", err)
	}
	return tok, nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

func saveToken(file string, token *oauth2.Token) error {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(token); err != nil {
		return err
	}
	return nil
}

func SetUpCredentials(ctx context.Context, clientSecretFilePath string, tokenCacheFilePath string) error {

	b, err := os.ReadFile(clientSecretFilePath)
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, youtube.YoutubeScope)
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	tok, err := getTokenFromWeb(ctx, config)
	if err != nil {
		return fmt.Errorf("unable to get token from web: %v", err)
	}
	if err := saveToken(tokenCacheFilePath, tok); err != nil {
		return fmt.Errorf("unable to save token: %v", err)
	}
	return nil
}

func NewYoutubeService(ctx context.Context, clientSecretFilePath string, tokenCacheFilePath string) (*youtube.Service, error) {
	b, err := os.ReadFile(clientSecretFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, youtube.YoutubeScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	tok, err := tokenFromFile(tokenCacheFilePath)
	if err != nil {
		return nil, err
	}
	client := config.Client(ctx, tok)
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	return service, err
}
