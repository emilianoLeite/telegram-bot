# LLM Adapter Usage

This Telegram bot now supports multiple LLM providers through an adapter pattern. You can choose between Hugging Face and Google AI Studio.

## Environment Variables

Set the following environment variables:

### Required for all adapters:
- `TELEGRAM_BOT_API_KEY`: Your Telegram bot API key
- `HUGGINGFACEHUB_API_TOKEN`: Your Hugging Face API token
- `GOOGLE_STUDIO_API_KEY`: Your Google AI Studio API key

### Adapter Selection:
- `LLM_ADAPTER_TYPE`: Choose between `huggingface` or `google_studio` (defaults to `huggingface` if not set)

## Usage Examples

### Using Hugging Face (default):
```bash
export TELEGRAM_BOT_API_KEY="your_telegram_key"
export HUGGINGFACEHUB_API_TOKEN="your_huggingface_token"
export GOOGLE_STUDIO_API_KEY="your_google_key"
# LLM_ADAPTER_TYPE defaults to huggingface
./telegram-bot
```

### Using Google AI Studio:
```bash
export TELEGRAM_BOT_API_KEY="your_telegram_key"
export HUGGINGFACEHUB_API_TOKEN="your_huggingface_token"
export GOOGLE_STUDIO_API_KEY="your_google_key"
export LLM_ADAPTER_TYPE="google_studio"
./telegram-bot
```

## Adapter Features

### Hugging Face Adapter
- Uses `facebook/blenderbot-400M-distill` model by default
- Supports text generation
- Image processing is not fully implemented (placeholder)

### Google AI Studio Adapter
- Uses `gemini-2.5-flash` model by default
- Supports text generation
- Supports image processing (multimodal)

## Architecture

The adapter pattern allows you to easily switch between different LLM providers without changing the main bot logic. Each adapter implements the `LLMAdapter` interface:

```go
type LLMAdapter interface {
    GenerateText(ctx context.Context, input string) (string, error)
    GenerateTextWithImage(ctx context.Context, input string, imageData []byte, mimeType string) (string, error)
}
```

## Adding New Adapters

To add a new LLM provider:

1. Create a new package in the `adapters/` directory
2. Implement the `LLMAdapter` interface
3. Add the new adapter type to the `AdapterType` enum
4. Update the factory function to handle the new adapter type
5. Update the initialization logic in `main.go`
