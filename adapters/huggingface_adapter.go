package adapters

import (
	"context"
	"fmt"

	huggingface "github.com/emilianoleite/telegram-bot/go-huggingface"
)

// HuggingFaceAdapter implements the LLMAdapter interface for Hugging Face
type HuggingFaceAdapter struct {
	client *huggingface.InferenceClient
	model  string
}

// NewHuggingFaceAdapter creates a new Hugging Face adapter
func NewHuggingFaceAdapter(token string, model string) (*HuggingFaceAdapter, error) {
	if token == "" {
		return nil, fmt.Errorf("huggingface token is required")
	}

	client := huggingface.NewInferenceClient(token)

	if model == "" {
		model = "facebook/blenderbot-400M-distill" // Default model
	}

	return &HuggingFaceAdapter{
		client: client,
		model:  model,
	}, nil
}

// GenerateText generates a text response based on the input text
func (h *HuggingFaceAdapter) GenerateText(ctx context.Context, input string) (string, error) {
	res, err := h.client.Text2TextGeneration(ctx, &huggingface.Text2TextGenerationRequest{
		Inputs: input,
		Model:  h.model,
		Parameters: huggingface.Text2TextGenerationParameters{
			Temperature: huggingface.PTR(0.7),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate text with Hugging Face: %w", err)
	}

	if len(res) == 0 {
		return "", fmt.Errorf("no response generated from Hugging Face")
	}

	return res[0].GeneratedText, nil
}

// GenerateTextWithImage generates a text response based on input text and image data
// Note: The current Hugging Face implementation doesn't support image input in this way
// This is a placeholder implementation that ignores the image
func (h *HuggingFaceAdapter) GenerateTextWithImage(ctx context.Context, input string, imageData []byte, mimeType string) (string, error) {
	// For now, we'll just use the text input and ignore the image
	// In a real implementation, you might want to use a vision model or multimodal model
	return h.GenerateText(ctx, input)
}

// Ensure HuggingFaceAdapter implements the LLMAdapter interface
var _ LLMAdapter = (*HuggingFaceAdapter)(nil)
