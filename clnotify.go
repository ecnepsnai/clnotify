// Command clnotify is an application to search craigslist and send matching postings to a Discord webhook
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ecnepsnai/craigslist"
	"github.com/ecnepsnai/discord"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config file path>\n", os.Args[0])
		os.Exit(1)
	}

	config, err := loadConfig(os.Args[1])
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	discord.WebhookURL = config.Discord.WebhookURL
	loadCache()

	params := craigslist.LocationParams{
		AreaID:         config.Craigslist.AreaID,
		Latitude:       config.Craigslist.Latitude,
		Longitude:      config.Craigslist.Longitude,
		SearchDistance: config.Craigslist.SearchDistance,
	}
	for _, search := range config.Searches {
		firstForSearch := cache.IsFirstForSearch(search.Category + search.Query)
		cache.SetFirstForSearch(search.Category + search.Query)

		results, err := craigslist.Search(search.Category, search.Query, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting results: query='%s' category='%s'\n", search.Query, search.Category)
			continue
		}

		for _, result := range results {
			if resultIsIgnored(result.Title, search.Ignore) {
				continue
			}

			if firstForSearch {
				cache.AddPost(result)
				continue
			}

			if !cache.HaveSeenPost(result) {
				discordPost(result, search.Name)
				cache.AddPost(result)
			}
		}
	}

	cache.store.Close()
}

func resultIsIgnored(resultTitle string, ignoredWords []string) bool {
	t := strings.ToLower(resultTitle)
	for _, w := range ignoredWords {
		if strings.Contains(t, strings.ToLower(w)) {
			return true
		}
	}

	return false
}

func discordPost(result craigslist.Result, searchName string) {
	post, err := result.Posting()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting post details: posting_id=%d error='%s'", result.PostingID, err.Error())
		return
	}

	message := discord.PostOptions{
		Content: "New Post for _" + searchName + "_",
		Embeds: []discord.Embed{
			{
				Title: result.Title,
				URL:   post.URL,
				Fields: []discord.Field{
					{
						Name:  "Price",
						Value: fmt.Sprintf("$%d", result.Price),
					},
				},
			},
		},
	}
	if len(result.Images) > 0 {
		url := result.ImageURLs()[0]
		image := discord.Image{
			URL: url,
		}
		message.Embeds[0].Image = &image
	}

	discord.Post(message)
}
