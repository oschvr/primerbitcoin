package notifications

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"primerbitcoin/pkg/utils"

	"os"
)

// SendTelegramMessage to send a message to a custom channel/conversation
func SendTelegramMessage(message string) {
	// Create a new bot using your API token.
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_API"))
	if err != nil {
		utils.Logger.Panic(err)
	}

	// Create a new message configuration.
	msg := tgbotapi.NewMessage(1386479921, message)

	// Use the bot.Send method to send the message.
	_, err = bot.Send(msg)
	if err != nil {
		log.Panic(err)
	}
}
