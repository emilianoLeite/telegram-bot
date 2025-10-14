# Adapter Usage Guide

This Telegram bot uses an adapter pattern to support multiple providers for both LLM processing and output destinations. This allows you to easily switch between different services without changing the core bot logic.

## Adapters Overview

### 1. LLM Adapters
Choose your AI provider for processing text and images (receipts).

### 2. Output Adapters
Choose where to send the processed receipt data.

## Environment Variables

- Check the [README.md](README.md) for a list of all required and optional env vars.


## LLM Adapter Features

### Hugging Face Adapter
- Uses `facebook/blenderbot-400M-distill` model by default
- Supports text generation
- Image processing is not fully implemented (placeholder)

### Google AI Studio Adapter (Recommended for Receipt Processing)
- Uses `gemini-2.5-flash` model by default
- Supports text generation
- Supports image processing (multimodal)

## Output Adapter Features

### Telegram Output Adapter (Default)
- Sends processed receipt data back to the user as a Telegram message
- Reply-to-message format for easy conversation tracking
- No additional configuration needed

### Google Sheets Output Adapter
- Appends receipt data to a Google Spreadsheet
- Requires service account credentials (`credentials.json`)
- Configurable sheet range (default: `Receipts!A:A`)
- Useful for tracking expenses over time

### File Output Adapter
- Saves receipt data to local text files
- Creates timestamped files (e.g., `receipt_2025-10-13_15-30-45.txt`)
- Useful for local archiving and processing


## Receipt Processing Flow

1. User sends a photo to the Telegram bot
2. Bot downloads the image from Telegram servers
3. image/jpeg MIME type is validated
4. **LLM Adapter** processes the image
5. **Output Adapter** writes the parsed receipt data to the configured destination
6. User receives confirmation or the processed data

## Adding New Adapters

### Adding a New LLM Provider:

1. Create a new package in the `adapters/` directory
2. Implement the `LlmAdapter` interface
3. Add the new adapter type to the `LlmAdapterType` enum in `llm_adapter.go`
4. Update the `NewLlmAdapter` factory function in `factory.go`
5. Update initialization logic in `main.go`

### Adding a New Output Destination:

1. Create a new file in the `adapters/` directory (e.g., `database_output.go`)
2. Implement the `OutputAdapter` interface
3. Add the new output type to the `OutputType` enum in `output_adapter.go`
4. Update the `NewOutputAdapter` factory function in `factory.go`
5. Add configuration handling in `main.go`
