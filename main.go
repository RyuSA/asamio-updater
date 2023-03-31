package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"

	"github.com/ryusa/asamio-updater/internal/auth"
	"github.com/ryusa/asamio-updater/internal/discord"
	"golang.org/x/net/context"
	"google.golang.org/api/youtube/v3"
)

var (
	DEFAULT_CLIENT_SECRET_FILE_PATH = ".credentials/client_secret.json"
	DEFAULT_TOKEN_FILE_PATH         = ".credentials/youtube.json"
)

func handleError(webhook *discord.DiscordWebhook, err error, message string) {
	if message == "" {
		message = "this is default msg"
	}
	if err == nil {
		return
	}

	if e := webhook.Do(discord.NewDiscordPayload(message + ": " + err.Error())); e != nil {
		log.Printf(message+": %v\n", err.Error())
		log.Fatalf("failed to send error message to discord: %v", err.Error())
	}
	log.Fatalf(message+": %v", err.Error())
}

func ListAllVideosInPlaylist(ctx context.Context, service *youtube.Service, playlistId string) ([]*youtube.PlaylistItem, error) {
	playlistCall := service.PlaylistItems.List([]string{"id", "snippet"}).PlaylistId(playlistId)

	result := make([]*youtube.PlaylistItem, 0)
	err := playlistCall.Pages(ctx, func(response *youtube.PlaylistItemListResponse) error {
		result = append(result, response.Items...)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func FilterVideos(service *youtube.Service, keyword string, channelId string) ([]*youtube.SearchResult, error) {
	r := regexp.MustCompile(keyword)
	call := service.Search.List([]string{"id", "snippet"})
	call = call.ChannelId(channelId).MaxResults(50).Order("date")
	response, err := call.Do()

	if err != nil {
		return nil, err
	}

	result := make([]*youtube.SearchResult, 0, len(response.Items))
	for _, item := range response.Items {
		// check if the video title contains the keyword with regular expression
		if r.MatchString(item.Snippet.Title) {
			result = append(result, item)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no videos matched the keyword %s", keyword)
	}
	return result, nil
}

func InsertVideoIntoPlaylist(service *youtube.Service, playlistId string, resourceId *youtube.ResourceId) error {
	call := service.PlaylistItems.Insert([]string{"snippet"}, &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistId,
			ResourceId: resourceId,
		},
	})
	if _, err := call.Do(); err != nil {
		return err
	}
	return nil
}

func main() {

	var (
		mioChannelId         string
		asamioPlayListId     string
		webhookEndpoint      string
		clientSecretFilePath string
		tokenCacheFilePath   string
		phase                string
	)
	flag.StringVar(&mioChannelId, "channelid", "", "channel id of @ookamimio")
	flag.StringVar(&asamioPlayListId, "playlistid", "", "playlist id of #朝ミオ")
	flag.StringVar(&webhookEndpoint, "webhook", "", "Discord Webhook endpoint")
	flag.StringVar(&clientSecretFilePath, "clientsecret", DEFAULT_CLIENT_SECRET_FILE_PATH, "client secret file path")
	flag.StringVar(&tokenCacheFilePath, "token", DEFAULT_TOKEN_FILE_PATH, "token cache file path")
	flag.StringVar(&phase, "phase", "update", "phase of the application, `init` or `update`")
	flag.Parse()

	ctx := context.Background()
	if phase == "init" {
		fmt.Println("initializing the YouTube token")
		if err := auth.SetUpCredentials(ctx, clientSecretFilePath, tokenCacheFilePath); err != nil {
			log.Fatalf("Unable to retrieve token from web: %v", err)
		}
		fmt.Println("done")
		return
	}

	webhook := discord.NewDiscordWebhook(webhookEndpoint)
	service, err := auth.NewYoutubeService(ctx, clientSecretFilePath, tokenCacheFilePath)
	if err != nil {
		handleError(webhook, err, "Error creating YouTube client")
	}

	asamioVideos, err := FilterVideos(service, "#朝ミオ", mioChannelId)
	if err != nil {
		handleError(webhook, err, fmt.Sprintf("Error filtering videos in channel %s", mioChannelId))
	}

	playlistVideos, err := ListAllVideosInPlaylist(ctx, service, asamioPlayListId)
	if err != nil {
		handleError(webhook, err, fmt.Sprintf("Error listing videos in playlist %s", asamioPlayListId))
	}

	updateCandidates := make([]*youtube.SearchResult, 0, len(asamioVideos))
	for _, asamioVideo := range asamioVideos {
		isExist := false
		for _, playlistVideo := range playlistVideos {
			if asamioVideo.Id.VideoId == playlistVideo.Snippet.ResourceId.VideoId {
				isExist = true
				break
			}
		}
		if !isExist {
			updateCandidates = append(updateCandidates, asamioVideo)
		}
	}

	fmt.Printf("candidates %d\n", len(updateCandidates))

	for index := range updateCandidates {
		candidate := updateCandidates[len(updateCandidates)-index-1]
		err := InsertVideoIntoPlaylist(service, asamioPlayListId, &youtube.ResourceId{
			Kind:    candidate.Id.Kind,
			VideoId: candidate.Id.VideoId,
		})
		if err != nil {
			handleError(webhook, err,
				fmt.Sprintf("Error inserting video %s into playlist %s", candidate.Id.VideoId, asamioPlayListId))
		}
	}

	if err := webhook.Do(discord.NewDiscordPayload(fmt.Sprintf("Update %d videos", len(updateCandidates)))); err != nil {
		handleError(webhook, err, fmt.Sprintf("Error sending webhook: %v", err))
	}

	fmt.Println("Done")
}
