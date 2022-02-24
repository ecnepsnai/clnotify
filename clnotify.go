// Command clnotify is an application to search craigslist and send matching postings to a Discord webhook
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ecnepsnai/craigslist"
	"github.com/ecnepsnai/discord"
	"github.com/ecnepsnai/logtic"
	"github.com/google/uuid"
)

var log = logtic.Log.Connect("clnotify")

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

	logtic.Log.FilePath = "clnotify.log"
	logtic.Log.Level = logtic.LevelError
	if config.Verbose {
		logtic.Log.Level = logtic.LevelDebug
	}
	logtic.Log.Open()
	defer logtic.Log.Close()

	discord.WebhookURL = config.Discord.WebhookURL
	loadCache()
	defer cache.store.Close()

	params := craigslist.LocationParams{
		AreaID:         config.Craigslist.AreaID,
		Latitude:       config.Craigslist.Latitude,
		Longitude:      config.Craigslist.Longitude,
		SearchDistance: config.Craigslist.SearchDistance,
	}
	for _, search := range config.Searches {
		for _, category := range search.Categories {
			log.PDebug("Running search", map[string]interface{}{
				"category": category,
				"query":    search.Query,
			})
			firstForSearch := cache.IsFirstForSearch(category + search.Query)
			if firstForSearch {
				log.Debug("First instance of search: %s", search.Name)
			}
			cache.SetFirstForSearch(category + search.Query)

			results, err := craigslist.Search(category, search.Query, params)
			if err != nil {
				log.PError("Error getting results", map[string]interface{}{
					"query":    search.Query,
					"category": category,
					"error":    err.Error(),
				})
				continue
			}
			log.PDebug("Search returned results", map[string]interface{}{
				"query":        search.Query,
				"category":     category,
				"result_count": len(results),
			})

			for _, result := range results {
				if resultIsIgnored(result.Title, search.Ignore) {
					log.PDebug("Listing is ignored", map[string]interface{}{
						"title": result.Title,
					})
					continue
				}

				if firstForSearch {
					cache.AddPost(result)
					continue
				}

				if !cache.HaveSeenPost(result) {
					log.PDebug("New post found", map[string]interface{}{
						"query":    search.Query,
						"category": category,
						"title":    result.Title,
					})
					discordPost(result, search.Name)
					cache.AddPost(result)
				}
			}
		}
	}
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
		log.PError("Error getting post details", map[string]interface{}{
			"posting_id": result.PostingID,
			"error":      err.Error(),
		})
		return
	}

	content := discord.PostOptions{
		Content: fmt.Sprintf("New listing for \"%s\": ($%d) **%s**: %s", searchName, result.Price, result.Title, post.URL),
	}

	if len(result.Images) > 0 {
		req, err := http.NewRequest("GET", result.ImageURLs()[0], nil)
		if err != nil {
			log.PError("Error forming HTTP request", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
		req.Header.Add("Pragma", "no-cache")
		req.Header.Add("Cache-Control", "no-cache")
		req.Header.Add("Upgrade-Insecure-Requests", "1")
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36 Edg/98.0.1108.56")
		req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		req.Header.Add("Accept-Encoding", "gzip, deflate, br")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.PError("Error getting post image", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		discord.UploadFile(content, discord.FileOptions{
			FileName: uuid.NewString() + ".jpg",
			Reader:   resp.Body,
		})
		resp.Body.Close()
	} else {
		discord.Post(content)
	}

	log.Info("%s", content.Content)
}
