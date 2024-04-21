package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const CAT_API = "https://api.thecatapi.com/v1/images/search?mime_types=jpg"
const N = 10

type Cache map[string]bool

func main() {
	urlChan := getImageUrls()
	fmt.Println("hallo")
	// TODO use in disk cache
	i := 0
	for url := range urlChan {
		fmt.Println(url)
		i++
	}
	fmt.Println("successfuly fetched", i, "urls")
}
func getImageUrls() <-chan string {
	var cache Cache = make(map[string]bool)
	out := make(chan string)
	wg := sync.WaitGroup{}

	action := func() {
		defer wg.Done()
		url, _ := getImageUrl(cache)
		if len(url) > 1 {
			out <- url
		}

	}

	for i := 0; i < N; i++ {
		wg.Add(1)
		go action()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out

}
func getImageUrl(cache Cache) (string, error) {
	response, err := http.Get(CAT_API)
	if err != nil {
		fmt.Printf("error getting image res %s\n", err)
		return "", err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("error getting image body %s\n", err)
		return "", err
	}
	var data [](map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Printf("error parsing json  %s\n", err)
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
	//fmt.Printf("%v %v", id, imgUrl)
	return imgUrl, nil
}
