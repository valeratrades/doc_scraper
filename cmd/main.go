package main

//TODO!: \
//Also most of these exchanges use cloudflare for caching, so pass some random query string at the end of url so it doesn't load from cloudflare's cache which could be stale.
//Learned it the hard way few years back.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Valera6/doc_scraper/utils"
)

type Hashes map[string]string

func getSHA256Hash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

func loadHashes(filePath string) (Hashes, error) {
	var hashes Hashes
	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(Hashes), nil
		}
		return nil, err
	}
	err = json.Unmarshal(file, &hashes)
	if err != nil {
		return nil, err
	}
	return hashes, nil
}

func saveHashes(filePath string, hashes Hashes) error {
	file, err := json.MarshalIndent(hashes, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, file, 0644)
}

func checkForChanges(hashes Hashes, key string, debug bool) {
	parts := strings.Split(key, "\n\n###\n\n")
	if len(parts) != 2 {
		fmt.Printf("Key format is incorrect, expecting 'url\\n\\n###\\n\\nhtmlClass' in hashes json file. Got: %s\n", key)
		return
	}
	url, htmlClass := parts[0], parts[1]
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch content from %s. Skipping...\n", url)
		return
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing the HTML from %s. Skipping...\n", url)
		return
	}
	contentBlock := ""
	doc.Find(htmlClass).Each(func(i int, s *goquery.Selection) {
		contentBlock += s.Text()
	})

	if debug {
		newlineCount := strings.Count(contentBlock, "\n")
		fmt.Printf("\nNumber of newlines in contentBlock for URL %s: %d\n", url, newlineCount)
	}

	newHash := getSHA256Hash(contentBlock)
	oldHash := hashes[key]
	if oldHash == "" || oldHash != newHash {
		fmt.Printf("Content changed for URL: %s\n", url)
		utils.MsgToValera(fmt.Sprintf("Content changed for URL: %s\n", url))
		hashes[key] = newHash
	}

}

func main() {
	filePath := "./data/hashes.json"
	debug := false

	hashes, err := loadHashes(filePath)
	if err != nil {
		panic(err)
	}
	for key := range hashes {
		checkForChanges(hashes, key, debug)
	}
	err = saveHashes(filePath, hashes)
	if err != nil {
		panic(err)
	}
}
