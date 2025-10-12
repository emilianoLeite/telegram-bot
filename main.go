package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/emilianoleite/telegram-bot/adapters"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	bot                   *tgbotapi.BotAPI
	telegram_api_key      string
	huggingface_token     string
	google_studio_api_key string
	llm_adapter           adapters.LLMAdapter
	adapter_type          adapters.AdapterType
)

func initializeLLMAdapter(adapterType adapters.AdapterType, model string) (adapters.LLMAdapter, error) {
	var apiKey string
	var env_var_exists bool

	switch adapterType {
	case adapters.HuggingFace:
		apiKey, env_var_exists = os.LookupEnv("HUGGINGFACEHUB_API_TOKEN")

		if apiKey == "" || !env_var_exists {
			log.Panic("Missing or invalid HUGGINGFACEHUB_API_TOKEN")
		}

		if model == "" {
			model = "facebook/blenderbot-400M-distill"
		}

		log.Printf("Initializing Hugging Face adapter with model: %s", model)
	case adapters.GoogleStudio:
		apiKey, env_var_exists = os.LookupEnv("GOOGLE_STUDIO_API_KEY")

		if apiKey == "" || !env_var_exists {
			log.Panic("Missing or invalid GOOGLE_STUDIO_API_KEY")
		}

		if model == "" {
			model = "gemini-2.5-flash"
		}
		log.Printf("Initializing Google AI Studio adapter with model: %s", model)
	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", adapterType)
	}

	return adapters.NewLLMAdapter(adapterType, apiKey, model)
}

func main() {
	var err error
	var env_var_exists bool

	telegram_api_key, env_var_exists = os.LookupEnv("TELEGRAM_BOT_API_KEY")

	if telegram_api_key == "" || !env_var_exists {
		log.Panic("Missing or invalid TELEGRAM_BOT_API_KEY")
	}

	google_studio_api_key, env_var_exists = os.LookupEnv("GOOGLE_STUDIO_API_KEY")

	if google_studio_api_key == "" || !env_var_exists {
		log.Panic("Missing or invalid GOOGLE_STUDIO_API_KEY")
	}

	// Get adapter type from environment variable, default to huggingface
	adapter_type_str, env_var_exists := os.LookupEnv("LLM_ADAPTER_TYPE")

	if !env_var_exists || adapter_type_str == "" {
		adapter_type = adapters.HuggingFace
		log.Println("LLM_ADAPTER_TYPE not set, defaulting to huggingface")
	} else {
		adapter_type = adapters.AdapterType(adapter_type_str)
	}

	llm_adapter, err = initializeLLMAdapter(adapter_type, "")

	if err != nil {
		log.Panicf("Failed to initialize LLM adapter: %v", err)
	}

	bot, err = tgbotapi.NewBotAPI(telegram_api_key)

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
	}
}

// When we get a command, we react accordingly
func handleCommand(_chatId int64, command string) error {
	var err error

	switch command {
	case "/start":
		log.Println("/start command received")
	}

	return err
}

func handleWithLLM(text string) (string, error) {
	if llm_adapter == nil {
		return "", fmt.Errorf("LLM adapter not initialized")
	}

	return llm_adapter.GenerateText(context.Background(), text)
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

		// Get the highest resolution photo
		photoSize := photo[len(photo)-1]

		file, err := bot.GetFile(tgbotapi.FileConfig{FileID: photoSize.FileID})
		if err != nil {
			log.Printf("Failed to get file: %s", err.Error())
			msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't process the image. Please try again.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			return
		}

		fileURL := file.Link(telegram_api_key)
		re := regexp.MustCompile(`\.([a-zA-Z0-9]+)(?:\?|$)`)

		matches := re.FindStringSubmatch(fileURL)

		if len(matches) > 1 {
			fileExt := matches[1]
			if fileExt != "jpg" {
				log.Printf("Unsupported file extension: %s", fileExt)
				msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't handle the provided file. Please provide an image.")
				msg.ReplyToMessageID = message.MessageID
				bot.Send(msg)
				return
			}
		} else {
			log.Printf("Could not detect file extension from URL: %s", fileURL)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't handle the provided file. Please provide an image.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			return
		}

		log.Printf("Downloading image from: %s", fileURL)
		resp, err := http.Get(fileURL)
		if err != nil {
			log.Printf("Failed to download image: %s", err.Error())
			msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't download the image. Please try again.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to download image. HTTP %s\nBody: %s", resp.Status, resp.Body)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't download the image. Please try again.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			return
		}
		defer resp.Body.Close()

		// Read image data
		imageData, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read image data: %s", err.Error())
			msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't read the image data. Please try again.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			return
		}

		log.Printf("Image data: %d bytes", len(imageData))

		// Process the image with the LLM
		prompt := "read this supermarket receipt and return the list of items in it"
		// Mime type is extracted from telegram file URL
		llmResponse, err := llm_adapter.GenerateTextWithImage(context.Background(), prompt, imageData, "image/jpeg")
		if err != nil {
			log.Printf("Failed to process image with LLM: %s", err.Error())
			msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't process the image. Please try again.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			return
		}

		// Send the LLM response
		msg := tgbotapi.NewMessage(message.Chat.ID, llmResponse)
		msg.ReplyToMessageID = message.MessageID
		_, err = bot.Send(msg)

		if err != nil {
			log.Printf("Failed to send message: %s", err.Error())
		}
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
