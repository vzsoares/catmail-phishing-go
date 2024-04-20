package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const CAT_API = "https://api.thecatapi.com/v1/images/search?mime_types=jpg"

type Cache map[string]bool

func main() {

	fmt.Println("hallo")
	// TODO use in disk cache
	var cache Cache = make(map[string]bool)
	for i := range 100 {
		getImageUrl(cache)
		fmt.Println(i)
	}
	getImageUrl(cache)
}
func getImageUrl(cache Cache) (string, error) {
	response, err := http.Get(CAT_API)
	if err != nil {
		fmt.Printf("error getting image res %s", err)
		return "", err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("error getting image body %s", err)
		return "", err
	}
	var data [](map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Printf("error parsing json  %s", err)
		return "", err
	}
	var imgUrl string = data[0]["url"].(string)
	var id string = data[0]["id"].(string)
	// check cache
	_, ok := cache[id]
	if ok {
		fmt.Printf("**already fetched**\n")
		return "", errors.New("already fetched")
	}
	cache[id] = true
	fmt.Printf("%v %v", id, imgUrl)
	return imgUrl, nil
}
