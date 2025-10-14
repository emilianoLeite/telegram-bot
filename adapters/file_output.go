package adapters

import (
	"context"
	"fmt"
	"os"
	"time"
)

type FileOutputAdapter struct {
	outputPath string
}

func NewFileOutputAdapter(outputPath string) (*FileOutputAdapter, error) {
	return &FileOutputAdapter{
		outputPath: outputPath,
	}, nil
}

func (f *FileOutputAdapter) WriteReceipt(ctx context.Context, receiptData string) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s/receipt_%s.txt", f.outputPath, timestamp)

	err := os.WriteFile(filename, []byte(receiptData), 0644)
	if err != nil {
		return fmt.Errorf("failed to write receipt to file: %w", err)
	}

	return nil
}
