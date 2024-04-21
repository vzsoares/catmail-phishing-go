package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

const CAT_API = "https://api.thecatapi.com/v1/images/search?mime_types=jpg"
const N = 100

type Cache struct {
	v  map[string]bool
	mu sync.Mutex
}

func main() {
	urlChan := getImageUrls()
	fmt.Println("hallo")
	pathChan := resolveImageToDisk(urlChan)

	i := 0
	for v := range pathChan {
		i++
        _ = v
	}
	fmt.Println("successfuly wrote", i, "images")
}
func resolveImageToDisk(c <-chan string) <-chan string {
	wg := sync.WaitGroup{}
	out := make(chan string)
	action := func(url string) {
		defer wg.Done()
		path, _ := downloadImgAndWriteToDisk(url)
		if len(path) > 1 {
			out <- path
		}
	}
	for url := range c {
		wg.Add(1)
		go action(url)
	}
	go func() {
		wg.Wait()
		defer close(out)
	}()
	return out
}
func getImageUrls() <-chan string {
	// TODO use in disk cache
	var cache Cache = Cache{v: make(map[string]bool)}
	out := make(chan string)
	wg := sync.WaitGroup{}
	action := func() {
		defer wg.Done()
		url, _ := getImageUrl(&cache)
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

func getImageUrl(cache *Cache) (string, error) {
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
	cache.mu.Lock()
	defer cache.mu.Unlock()
	_, ok := cache.v[id]
	if ok {
		fmt.Printf("**already fetched**\n")
		return "", errors.New("already fetched")
	}
	cache.v[id] = true
	return imgUrl, nil
}

func downloadImgAndWriteToDisk(imgUrl string) (string, error) {
	const basePath = "image/"
	name := strings.Split(imgUrl, "/")
	pathName := basePath + name[len(name)-1]
	if _, err := os.Stat(pathName); err == nil {
		fmt.Println("image already created")
		return "", errors.New("already created")
	}

	response, err := http.Get(imgUrl)
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
	img, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		fmt.Printf("error decoding image  %s\n", err)
		return "", err
	}
	//fmt.Println(pathName)
	outputFile, err := os.Create(pathName)
	if err != nil {
		fmt.Printf("error os.Create  %s\n", err)
		return "", err
	}
	defer outputFile.Close()
	err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: 100})
	if err != nil {
		fmt.Printf("error writing img  %s\n", err)
		return "", err
	}
	return pathName, nil
}
