package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hupe1980/go-huggingface"
)

var (
	bot               *tgbotapi.BotAPI
	api_key           string
	huggingface_token string
)

func main() {
	var err error
	var env_var_exists bool

	api_key, env_var_exists = os.LookupEnv("TELEGRAM_BOT_API_KEY")

	if api_key == "" || !env_var_exists {
		log.Panic("Missing or invalid TELEGRAM_BOT_API_KEY")
	}

	huggingface_token, env_var_exists = os.LookupEnv("HUGGINGFACEHUB_API_TOKEN")

	if huggingface_token == "" || !env_var_exists {
		log.Panic("Missing or invalid HUGGINGFACEHUB_API_TOKEN")
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

func handleWithLLM(text string) (string, error) {

	ic := huggingface.NewInferenceClient(huggingface_token)

	// res, err := ic.Conversational(context.Background(), &huggingface.ConversationalRequest{
	// 	Inputs: huggingface.ConverstationalInputs{
	// 		// PastUserInputs: []string{
	// 		// 	"Which movie is the best ?",
	// 		// 	"Can you explain why ?",
	// 		// },
	// 		// GeneratedResponses: []string{
	// 		// 	"It's Die Hard for sure.",
	// 		// 	"It's the best movie ever.",
	// 		// },
	// 		Text: text,
	// 	},
	// 	Model: "google/gemma-2-2b-it",
	// })

	res, err := ic.Text2TextGeneration(context.Background(), &huggingface.Text2TextGenerationRequest{
		Inputs: text,
		Model:  "facebook/blenderbot-400M-distill",
		Parameters: huggingface.Text2TextGenerationParameters{
			// Prevent the model from generating very long responses
			// MaxNewTokens: huggingface.PTR(100),
			// Add some randomness to responses
			Temperature: huggingface.PTR(0.7),
		},
	})

	if err != nil {
		return "", err
	}

	return res[0].GeneratedText, nil
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
			llm_response, err := handleWithLLM(text)
			if err != nil {
				log.Printf("[HuggingFace] Unexpected error: %s\n", err.Error())

				msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry I could not handle your message, please try again later")
				msg.ReplyToMessageID = message.MessageID
				_, err = bot.Send(msg)

				if err != nil {
					log.Printf("[Telegram] Failed to send message: %s\n", err.Error())
					return
				}
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, llm_response)
			msg.ReplyToMessageID = message.MessageID
			_, err = bot.Send(msg)

			if err != nil {
				log.Printf("[Telegram] Failed to send message: %s\n", err.Error())
				return
			}
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
			log.Printf("[Telegram] Failed to send message: %s\n", err.Error())
		}
	}
}
