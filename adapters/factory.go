package adapters

import (
	"fmt"

	"github.com/emilianoleite/telegram-bot/adapters/google_ai_studio"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewLlmAdapter(adapterType LlmAdapterType, apiKey string, model string) (LlmAdapter, error) {
	switch adapterType {
	case LLM.HuggingFace:
		return NewHuggingFaceAdapter(apiKey, model)
	case LLM.GoogleStudio:
		return google_ai_studio.NewGoogleStudioAdapter(apiKey, model)
	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", adapterType)
	}
}

type OutputAdapterConfig struct {
	// For Google Sheets
	CredentialsJSON string
	SpreadsheetID   string
	SheetRange      string

	// For File
	OutputPath string

	// For Telegram
	Bot          *tgbotapi.BotAPI
	ChatID       int64
	ReplyToMsgID int
}

func NewOutputAdapter(outputType OutputAdapterType, config OutputAdapterConfig) (OutputAdapter, error) {
	switch outputType {
	case Output.GoogleSheets:
		return NewGoogleSheetsOutputAdapter(config.CredentialsJSON, config.SpreadsheetID, config.SheetRange)
	case Output.File:
		return NewFileOutputAdapter(config.OutputPath)
	case Output.Telegram:
		return NewTelegramOutputAdapter(config.Bot, config.ChatID, config.ReplyToMsgID)
	default:
		return nil, fmt.Errorf("unsupported output type: %s", outputType)
	}
}
