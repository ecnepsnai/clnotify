package main

import (
	"strings"

	"github.com/ecnepsnai/craigslist"
	"github.com/ecnepsnai/store"
)

type cacheType struct {
	store *store.Store
}

var cache cacheType

func loadCache() {
	d, err := store.New(".", "clnotifycache")
	if err != nil {
		panic(err)
	}

	cache = cacheType{store: d}
}

func (c cacheType) IsFirstForSearch(searchName string) bool {
	return len(c.store.Get("search_first_"+searchName)) == 0
}

func (c cacheType) SetFirstForSearch(searchName string) {
	c.store.Write("search_first_"+searchName, []byte("1"))
}

func (c cacheType) AddPost(result craigslist.Result) {
	c.store.Write("post_"+result.DedupeKey, []byte("1"))
	c.store.Write("title_"+normalizeTitle(result.Title), []byte("1"))
}

func (c cacheType) HaveSeenPost(result craigslist.Result) bool {
	key := len(c.store.Get("post_" + result.DedupeKey))
	title := len(c.store.Get("title_" + normalizeTitle(result.Title)))
	return key+title > 0
}

func normalizeTitle(title string) string {
	t := title
	t = strings.ReplaceAll(t, " ", "")
	t = strings.ReplaceAll(t, "\"", "")
	t = strings.ReplaceAll(t, "+", "")
	return strings.ToLower(t)
}
