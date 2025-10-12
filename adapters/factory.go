package adapters

import (
	"fmt"

	"github.com/emilianoleite/telegram-bot/adapters/google_ai_studio"
)

// NewLLMAdapter creates a new LLM adapter based on the adapter type
func NewLLMAdapter(adapterType AdapterType, apiKey string, model string) (LLMAdapter, error) {
	switch adapterType {
	case HuggingFace:
		return NewHuggingFaceAdapter(apiKey, model)
	case GoogleStudio:
		return google_ai_studio.NewGoogleStudioAdapter(apiKey, model)
	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", adapterType)
	}
}
