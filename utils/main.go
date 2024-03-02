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

func MsgToValera(msg string) {
	const (
		valeraBotToken = "6225430873:AAEYlbJ2bY-WsLADxlWY1NS-z4r75sf9X5I"
		valeraChatID   = -1001800341082
	)

	Msg(valeraBotToken, valeraChatID, msg)
}
