# doc_scraper
Tiny util, intended to be run once a day, in order to detect any changes in api documentation.

# Installation
Need to build the application, then move the hashes file to a convenient location, then run `doc_scraper init` to rehash the defined there endpoints.
For example:
```sh
git clone --depth=1 https://github.com/Valera6/doc_scraper /tmp/doc_scraper && \
cd /tmp/doc_scraper && \
sudo go build -o /usr/local/bin/doc_scraper ./cmd/main.go && \
cd - &>/dev/null && \
mkdir -p ~/tmp && cp /tmp/doc_scraper/starting_hashes.json ~/tmp/doc_scraper_hashes.json && \
doc_scraper init
```

# Usage
After having had built and initialized, schedule it to be run once a day or so.
Command to run is:
```sh
doc_scraper check # optionally provide --path argument, if the hashes file is not in ~/tmp/doc_scraper_hashes.json
```
If any changes are detected:
    - prints them to stderr
    - sends message to a tg channel, if flag with (token,chatID) provided
    - exits with 1

# Limitations
- Made with Linux in mind.
- Currently working with Binance only. (easy to add others if needed - open an issue)
- Sends messages to a telegram channel only. (easy to add others - open an issue)
