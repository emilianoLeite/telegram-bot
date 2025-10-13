package adapters

import "context"

type LlmAdapter interface {
	// GenerateText generates a text response based on the input text
	GenerateText(ctx context.Context, input string) (string, error)

	// GenerateTextWithImage generates a text response based on input text and image data
	GenerateTextWithImage(ctx context.Context, input string, imageData []byte, mimeType string) (string, error)
}

type LlmAdapterType string

var LLM = struct {
	HuggingFace  LlmAdapterType
	GoogleStudio LlmAdapterType
}{
	HuggingFace:  "huggingface",
	GoogleStudio: "google_studio",
}
