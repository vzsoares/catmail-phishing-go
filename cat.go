package main

// TODO build email
// TODO send email
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"main/utils"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/h2non/bimg"
)

const CAT_API = "https://api.thecatapi.com/v1/images/search?mime_types=jpg"
const N = 10

type Cache struct {
	v  map[string]bool
	mu sync.Mutex
}

func main() {
	// startup
	initChan := make(chan int, N)
	for i := 0; i < N; i++ {
		initChan <- i
	}
	close(initChan)

	var cache Cache = Cache{v: make(map[string]bool)}

	urlChanAction := func(i int) string {
		fmt.Println("getting url")
		url, _ := getImageUrl(&cache)
		return url
	}
	urlChan := utils.ChainOrchestrator(initChan, urlChanAction)
	//
	pathChanAction := func(url string) string {
		fmt.Println("writing to disk")
		path, _ := downloadImgAndWriteToDisk(url)
		return path
	}
	pathChan := utils.ChainOrchestrator(urlChan, pathChanAction)
	//
	errChanAction := func(url string) error {
		fmt.Println("writing watermark")
		err := putWatermark(url)
		return err
	}
	errChan := utils.ChainOrchestrator(pathChan, errChanAction)

	i := 0
	for v := range errChan {
		i++
		_ = v
	}
	fmt.Println("successfuly wrote", i, "images")
}

// actions

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

func putWatermark(filePath string) error {
	watermarkBuff, err := bimg.Read("clickme.jpg")
	if watermarkBuff != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	buffer, err := bimg.Read(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	watermark := bimg.WatermarkImage{
		Left:    0,
		Top:     0,
		Opacity: 100,
		Buf:     watermarkBuff,
	}

	newImage, err := bimg.NewImage(buffer).WatermarkImage(watermark)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	bimg.Write(filePath, newImage)
	return nil
}
