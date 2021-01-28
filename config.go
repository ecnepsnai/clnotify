package main

import (
	"encoding/json"
	"os"
)

type clnotifyConfigType struct {
	Craigslist clnotifyConfigCraigslistType `json:"craigslist"`
	Discord    clnotifyConfigDiscordType    `json:"discord"`
	Searches   []clnotifyConfigSearchType   `json:"searches"`
}

type clnotifyConfigCraigslistType struct {
	AreaID         int     `json:"area_id"`
	Latitude       float32 `json:"latitude"`
	Longitude      float32 `json:"longitude"`
	SearchDistance int     `json:"search_distance"`
}

type clnotifyConfigDiscordType struct {
	WebhookURL string `json:"webhook_url"`
}

type clnotifyConfigSearchType struct {
	Category string   `json:"category"`
	Query    string   `json:"query"`
	Name     string   `json:"name"`
	Ignore   []string `json:"ignore"`
}

func loadConfig(filePath string) (*clnotifyConfigType, error) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	config := clnotifyConfigType{}
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
