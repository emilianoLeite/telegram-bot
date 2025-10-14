package adapters

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsOutputAdapter struct {
	service         *sheets.Service
	credentialsJSON string
	spreadsheetID   string
	sheetRange      string
}

func NewGoogleSheetsOutputAdapter(credentialsJSON, spreadsheetID, sheetRange string) (*GoogleSheetsOutputAdapter, error) {
	return &GoogleSheetsOutputAdapter{
		credentialsJSON: credentialsJSON,
		spreadsheetID:   spreadsheetID,
		sheetRange:      sheetRange,
	}, nil
}

// WriteReceipt writes receipt data to a Google Sheet.
// receiptData should be a single string, e.g. "Item: Coffee, Price: $3.50, Date: 2023-09-01"
// To write multiple cells per row, separate values with a comma and split like so:
//
//	items := strings.Split(receiptData, ",")
//	values := []SheetRow{items}
func (g *GoogleSheetsOutputAdapter) WriteReceipt(ctx context.Context, receiptData string) error {
	// Initialize service if not already done
	if g.service == nil {
		if err := g.initService(ctx); err != nil {
			return fmt.Errorf("failed to initialize service: %w", err)
		}
	}

	// Example: receiptData should be a single string, e.g. "Item: Coffee, Price: $3.50, Date: 2023-09-01"
	// To write multiple cells per row, separate values with a comma and split like so:
	// items := strings.Split(receiptData, ",")
	// values := [][]interface{}{items}
	values := [][]interface{}{
		{receiptData},
	}

	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := g.service.Spreadsheets.Values.Append(
		g.spreadsheetID,
		g.sheetRange,
		valueRange,
	).ValueInputOption("USER_ENTERED").Do()

	if err != nil {
		return fmt.Errorf("unable to write data to sheet: %w", err)
	}

	return nil
}

// initService initializes the Google Sheets service
func (g *GoogleSheetsOutputAdapter) initService(ctx context.Context) error {
	config, err := google.JWTConfigFromJSON([]byte(g.credentialsJSON), "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return fmt.Errorf("unable to parse credentials: %w", err)
	}

	client := config.Client(ctx)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to create Sheets client: %w", err)
	}

	g.service = srv
	return nil
}
