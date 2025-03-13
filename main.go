package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	bot *tgbotapi.BotAPI
)

func main() {
	var err error
	api_key, exists := os.LookupEnv("TELEGRAM_BOT_API_KEY")

	if api_key == "" || !exists {
		log.Panic("Missing or invalid TELEGRAM_BOT_API_KEY")
	}

	bot, err = tgbotapi.NewBotAPI(api_key)

	if err != nil {
		// Abort if something is wrong
		log.Panic(err)
	}

	// Set this to true to log all interactions with telegram servers
	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Create a new cancellable background context. Calling `cancel()` leads to the cancellation of the context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// `updates` is a golang channel which receives telegram updates
	updates := bot.GetUpdatesChan(u)

	// Pass cancellable context to goroutine
	go receiveUpdates(ctx, updates)

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press enter to stop")

	// Wait for a newline symbol, then cancel handling updates
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()
}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	// `for {` means the loop is infinite until we manually stop it
	for {
		select {
		// stop looping if ctx is cancelled
		case <-ctx.Done():
			return
		// receive update from channel and then handle it
		case update := <-updates:
			handleUpdate(update)
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		handleMessage(update.Message)
		break
	}
}

// When we get a command, we react accordingly
func handleCommand(_chatId int64, command string) error {
	var err error

	switch command {
	case "/start":
		log.Println("/start command received")
		break
	}

	return err
}

func handleMessage(message *tgbotapi.Message) {
	user := message.From
	text := message.Text
	photo := message.Photo
	var err error

	if user == nil {
		log.Println("User is nil, message rejected")
		return
	}

	if text != "" {
		// Print to console
		log.Printf("Received message: %+v", text)

		if strings.HasPrefix(text, "/") {
			err = handleCommand(message.Chat.ID, text)
		} else {
			// This is equivalent to forwarding, without the sender's name
			copyMsg := tgbotapi.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
			_, err = bot.CopyMessage(copyMsg)
		}

		if err != nil {
			log.Printf("An error occured: %s", err.Error())
		}
	} else if len(photo) > 0 {
		log.Println("Handling photo message")
		// msg := tgbotapi.NewMessage(message.Chat.ID, "Success")
		// msg.ReplyToMessageID = message.MessageID
		// _, err = bot.Send(msg)

		// if err != nil {
		// 	log.Printf("An error occured: %s", err.Error())
		// }
	} else {
		log.Println("Unexpected message type, message rejected")
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unexpected message type, message rejected")
		msg.ReplyToMessageID = message.MessageID
		_, err = bot.Send(msg)

		if err != nil {
			log.Printf("An error occured: %s", err.Error())
		}
	}

}
