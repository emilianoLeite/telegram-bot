package adapters

import "context"

// LLMAdapter defines the interface for different LLM providers
type LLMAdapter interface {
	// GenerateText generates a text response based on the input text
	GenerateText(ctx context.Context, input string) (string, error)

	// GenerateTextWithImage generates a text response based on input text and image data
	GenerateTextWithImage(ctx context.Context, input string, imageData []byte, mimeType string) (string, error)
}

// AdapterType represents the type of LLM adapter
type AdapterType string

const (
	HuggingFace  AdapterType = "huggingface"
	GoogleStudio AdapterType = "google_studio"
)
