package main

import (
	"fmt"
	torrentapi "github.com/ricksancho/rarbg-torrentapi"
	"time"
)

func main() {
	t, err := torrentapi.New(1337)
	if err != nil {
		panic(err)
	}
	err = t.Init()
	if err != nil {
		panic(err)
	}

	var results torrentapi.TorrentResults
	results, err = t.List(map[string]string{"category": "movies", "sort": "seeders"})
	if err != nil {
		panic(err)
	}
	fmt.Println(len(results.Torrents))
	time.Sleep(time.Duration(2) * time.Second)

	results, err = t.Search(map[string]string {"search_string": "Westworld season 1", "sort": "seeders"})
	if err != nil {
		panic(err)
	}

	fmt.Println(len(results.Torrents))
}
