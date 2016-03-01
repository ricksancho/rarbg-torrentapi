package main

import (
	"fmt"
	torrentapi "github.com/ricksancho/rarbg-torrentapi"
	"time"
)

func main() {
	t, err := torrentapi.New()
	if err != nil {
		panic(err)
	}
	err = t.Init()
	if err != nil {
		panic(err)
	}

	err = t.List(map[string]string{"category": "movies", "sort": "seeders"})
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(time.Duration(2) * time.Second)

	for i := 0; i < 2; i++ {
		err = t.Search(map[string]string{
			"search_string": "Shaun of the dead",
			"sort":          "seeders"})
		/*err = t.Search(map[string]string{
		"search_imdb": "tt1979388",
		"sort":        "seeders"})*/
		/*err = t.Search(map[string]string{
		"search_imdb":   "tt4016454",
		"search_string": "S01E10", "sort": "seeders"})*/
		if err != nil {
			fmt.Println(err)
		}
	}

}
