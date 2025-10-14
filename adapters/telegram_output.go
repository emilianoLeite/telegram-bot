package adapters

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramOutputAdapter implements the OutputAdapter interface for Telegram
type TelegramOutputAdapter struct {
	bot          *tgbotapi.BotAPI
	chatID       int64
	replyToMsgID int
}

// NewTelegramOutputAdapter creates a new Telegram output adapter
func NewTelegramOutputAdapter(bot *tgbotapi.BotAPI, chatID int64, replyToMsgID int) (*TelegramOutputAdapter, error) {
	return &TelegramOutputAdapter{
		bot:          bot,
		chatID:       chatID,
		replyToMsgID: replyToMsgID,
	}, nil
}

// WriteReceipt writes the parsed receipt data as a Telegram message
func (t *TelegramOutputAdapter) WriteReceipt(ctx context.Context, receiptData string) error {
	msg := tgbotapi.NewMessage(t.chatID, receiptData)
	msg.ReplyToMessageID = t.replyToMsgID

	_, err := t.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send receipt via Telegram: %w", err)
	}

	return nil
}
