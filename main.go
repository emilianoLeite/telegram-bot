package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/emilianoleite/telegram-bot/adapters"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	telegramAPIKey string
)

func initializeLlmAdapter(adapterType adapters.LlmAdapterType, model string) (adapters.LlmAdapter, error) {
	var apiKey string
	var envVarExists bool

	switch adapterType {
	case adapters.LLM.HuggingFace:
		apiKey, envVarExists = os.LookupEnv("HUGGINGFACEHUB_API_TOKEN")

		if apiKey == "" || !envVarExists {
			log.Panic("Missing or invalid HUGGINGFACEHUB_API_TOKEN")
		}

		if model == "" {
			model = "facebook/blenderbot-400M-distill"
		}

		log.Printf("Initializing Hugging Face adapter with model: %s", model)
	case adapters.LLM.GoogleStudio:
		apiKey, envVarExists = os.LookupEnv("GOOGLE_STUDIO_API_KEY")

		if apiKey == "" || !envVarExists {
			log.Panic("Missing or invalid GOOGLE_STUDIO_API_KEY")
		}

		if model == "" {
			model = "gemini-2.5-flash"
		}
		log.Printf("Initializing Google AI Studio adapter with model: %s", model)
	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", adapterType)
	}

	return adapters.NewLlmAdapter(adapterType, apiKey, model)
}

func main() {
	var err error
	var envVarExists bool

	telegramAPIKey, envVarExists := os.LookupEnv("TELEGRAM_BOT_API_KEY")

	if telegramAPIKey == "" || !envVarExists {
		log.Panic("Missing or invalid TELEGRAM_BOT_API_KEY")
	}

	llmAdapterType, envVarExists := os.LookupEnv("LLM_ADAPTER_TYPE")
	var adapterType adapters.LlmAdapterType

	if !envVarExists || llmAdapterType == "" {
		adapterType = adapters.LLM.HuggingFace
		log.Println("LLM_ADAPTER_TYPE not set, defaulting to huggingface")
	} else {
		adapterType = adapters.LlmAdapterType(llmAdapterType)
	}

	llmAdapter, err := initializeLlmAdapter(adapterType, "")

	if err != nil {
		log.Panicf("Failed to initialize LLM adapter: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(telegramAPIKey)

	if err != nil {
		log.Panic(err)
	}

	// Set this to true to log all interactions with telegram servers
	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	updatesChan := bot.GetUpdatesChan(u)

	go receiveUpdates(ctx, bot, llmAdapter, updatesChan)

	log.Println("Bot is now running. Press Ctrl+C to stop.")

	// Set up channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		os.Interrupt,    // Ctrl+C
		syscall.SIGTERM, // kill
		syscall.SIGQUIT, // Ctrl+\
	)

	sig := <-sigChan
	log.Printf("Received %v signal, shutting down...", sig)
}

func receiveUpdates(
	ctx context.Context, bot *tgbotapi.BotAPI, llmAdapter adapters.LlmAdapter, updates tgbotapi.UpdatesChannel,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			handleUpdate(ctx, bot, llmAdapter, update)
		}
	}
}

func handleUpdate(
	ctx context.Context, bot *tgbotapi.BotAPI, llmAdapter adapters.LlmAdapter, update tgbotapi.Update,
) {
	switch {
	case update.Message != nil:
		handleMessage(ctx, bot, llmAdapter, update.Message)
	}
}

func handleCommand(_chatId int64, command string) error {
	var err error

	switch command {
	case "/start":
		log.Println("/start command received")
	}

	return err
}

func handleMessage(
	ctx context.Context, bot *tgbotapi.BotAPI, llmAdapter adapters.LlmAdapter, message *tgbotapi.Message,
) {
	user := message.From
	text := message.Text
	photo := message.Photo
	var err error

	if user == nil {
		log.Println("User is nil, message rejected")
		return
	}

	if text != "" {
		log.Printf("Received message: %+v", text)

		if strings.HasPrefix(text, "/") {
			err = handleCommand(message.Chat.ID, text)
		} else {
			llmResponse, err := llmAdapter.GenerateText(ctx, text)
			if err != nil {
				log.Printf("[LLM] Unexpected error: %s\n", err.Error())

				msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry I could not handle your message, please try again later")
				msg.ReplyToMessageID = message.MessageID
				_, err = bot.Send(msg)

				if err != nil {
					log.Printf("[Telegram] Failed to send message: %s\n", err.Error())
					return
				}
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, llmResponse)
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

		fileURL := file.Link(telegramAPIKey) // TODO: refactor bot into custom class so we don't need to have telegramAPIKey as a global variable
		fileExtensionRegex := regexp.MustCompile(`\.([a-zA-Z0-9]+)(?:\?|$)`)
		matches := fileExtensionRegex.FindStringSubmatch(fileURL)

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

		imageData, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read image data: %s", err.Error())
			msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't read the image data. Please try again.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			return
		}

		log.Printf("Image data: %d bytes", len(imageData))

		prompt := "read this supermarket receipt and return the list of items in it"
		// Mime type is extracted from telegram file URL
		_llmResponse, err := llmAdapter.GenerateTextWithImage(ctx, prompt, imageData, "image/jpeg")
		if err != nil {
			log.Printf("Failed to process image with LLM: %s", err.Error())
			msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't process the image. Please try again.")
			msg.ReplyToMessageID = message.MessageID
			bot.Send(msg)
			return
		}

		// TODO: properly review all this 👇

		// outputTypeStr := os.Getenv("OUTPUT_TYPE")
		// if outputTypeStr == "" {
		// 	outputTypeStr = string(adapters.Output.File)
		// }
		// outputType := adapters.OutputAdapterType(outputTypeStr)

		// var outputConfig adapters.OutputAdapterConfig
		// switch outputType {
		// case adapters.Output.GoogleSheets:
		// 	credentialsJSON, _ := os.ReadFile("credentials.json")
		// 	outputConfig = adapters.OutputAdapterConfig{
		// 		CredentialsJSON: string(credentialsJSON),
		// 		SpreadsheetID:   os.Getenv("GOOGLE_SHEET_ID"),
		// 		SheetRange:      "Receipts!A:A",
		// 	}
		// case adapters.Output.File:
		// 	outputConfig = adapters.OutputAdapterConfig{
		// 		OutputPath: os.Getenv("OUTPUT_PATH"),
		// 	}
		// case adapters.Output.Telegram:
		// 	outputConfig = adapters.OutputAdapterConfig{
		// 		Bot:          bot,
		// 		ChatID:       message.Chat.ID,
		// 		ReplyToMsgID: message.MessageID,
		// 	}
		// }

		// outputAdapter, err := adapters.NewOutputAdapter(outputType, outputConfig)
		// if err != nil {
		// 	log.Printf("Failed to create output adapter: %s", err.Error())
		// 	msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't save the receipt. Please try again.")
		// 	msg.ReplyToMessageID = message.MessageID
		// 	bot.Send(msg)
		// 	return
		// }

		// err = outputAdapter.WriteReceipt(ctx, llmResponse)
		// if err != nil {
		// 	log.Printf("Failed to write receipt: %s", err.Error())
		// 	msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't save the receipt. Please try again.")
		// 	msg.ReplyToMessageID = message.MessageID
		// 	bot.Send(msg)
		// 	return
		// }

		// log.Printf("Receipt processed and written via %s adapter", outputType)
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
