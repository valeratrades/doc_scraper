package main

//TODO!: \
//Also most of these exchanges use cloudflare for caching, so pass some random query string at the end of url so it doesn't load from cloudflare's cache which could be stale.
//Learned it the hard way few years back.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Valera6/doc_scraper/utils"
	"github.com/urfave/cli"
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

func writeChanges(hashes Hashes, key string, init bool, tgArgs TgArgs) {
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

	if init {
		newlineCount := strings.Count(contentBlock, "\n")
		fmt.Printf("Number of newlines in contentBlock for URL %s: %d\n", url, newlineCount)
		return
	}

	newHash := getSHA256Hash(contentBlock)
	oldHash := hashes[key]
	if oldHash == "" || oldHash != newHash {
		fmt.Printf("Content changed for URL: %s\n", url)
		utils.Msg(tgArgs.BotToken, tgArgs.ChatId, fmt.Sprintf("Content changed for URL: %s\n", url))
		hashes[key] = newHash
	}
}

type TgArgs struct {
	BotToken string
	ChatId   int64
}

func NewTgArgs(input string) (TgArgs, error) {
	if input == "" {
		return TgArgs{
			BotToken: "6225430873:AAEYlbJ2bY-WsLADxlWY1NS-z4r75sf9X5I",
			ChatId:   -1001800341082,
		}, nil
	}

	parts := strings.Split(input, ",")
	if len(parts) != 2 {
		return TgArgs{}, fmt.Errorf("expected input format 'token,chatID', got: %s", input)
	}

	chatId, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return TgArgs{}, fmt.Errorf("invalid chat ID: %s", parts[1])
	}

	return TgArgs{
		BotToken: parts[0],
		ChatId:   chatId,
	}, nil
}

func runApplication(c *cli.Context) error {
	initFlag := c.Command.Name == "init"
	tgInfo := c.String("telegram")

	var tgArgs TgArgs
	var err error

	tgArgs, err = NewTgArgs(tgInfo)
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	defaultPath := filepath.Join(homeDir, "tmp", "doc_scraper_hashes.json")
	filePath := c.String("path")
	if filePath == "" {
		filePath = defaultPath
	}

	originalHashes, err := loadHashes(filePath)
	if err != nil {
		return err
	}
	hashes := make(Hashes, len(originalHashes))
	for k, v := range originalHashes {
		hashes[k] = v
	}
	for key := range hashes {
		writeChanges(hashes, key, initFlag, tgArgs)
	}
	err = saveHashes(filePath, hashes)
	if err != nil {
		return err
	}

	if !initFlag {
		for key := range hashes {
			if hashes[key] != originalHashes[key] {
				os.Exit(1)
			}
		}
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "doc_scraper"
	app.Usage = "Stupid little thing to catch exchange documentation changes."
	app.Commands = []cli.Command{
		{
			Name:   "check",
			Usage:  "Loads hashes and url:tmlClass from ./data/hashes.json",
			Action: runApplication,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "telegram",
					Usage: "Telegram bot token and chat ID to receive notification on; format: 'token,chatID'",
				},
				&cli.StringFlag{
					Name:  "path",
					Usage: "Path to the hashes.json file, default '~/tmp/doc_scraper_hashes.json'",
				},
			},
		},
		{
			Name:  "init",
			Usage: "Initialize the thing without spamming yourself",
			Action: func(c *cli.Context) error {
				return runApplication(c)
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "path",
					Usage: "Path to the hashes.json file, default '~/tmp/doc_scraper_hashes.json'",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
