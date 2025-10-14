package adapters

import "context"

type OutputAdapter interface {
	// WriteReceipt writes the parsed receipt data to the output destination
	WriteReceipt(ctx context.Context, receiptData string) error
}

type OutputAdapterType string

var Output = struct {
	GoogleSheets OutputAdapterType
	File         OutputAdapterType
	Telegram     OutputAdapterType
}{
	GoogleSheets: "google_sheets",
	File:         "file",
	Telegram:     "telegram",
}
