package utils

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func Msg(botToken string, chatID int64, msg string) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic("Failed to create bot:", err)
	}

	message := tgbotapi.NewMessage(chatID, msg)
	_, err = bot.Send(message)
	if err != nil {
		log.Println("Error sending message:", err)
	}
}
