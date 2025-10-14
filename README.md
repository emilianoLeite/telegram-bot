# Telegram Receipt Scanner Bot

A Telegram bot that processes supermarket receipts using AI and sends the extracted data to your preferred output destination (Telegram, Google Sheets, or local files).

## Features

- 📸 **Receipt Image Processing**: Send a photo of your receipt and get a parsed list of items
- 🤖 **Multiple LLM Providers**: Choose between Google AI Studio (Gemini) or Hugging Face
- 📊 **Flexible Output Options**: Send results via Telegram, save to Google Sheets, or write to local files
- 🔄 **Adapter Pattern Architecture**: Easy to extend with new providers and output destinations

## Prerequisites

- Go 1.24 or higher
- Telegram Bot API key
- One of the following LLM API keys:
  - Google AI Studio API key
  - Hugging Face API token
- (Optional) Google Cloud service account credentials for Google Sheets output

## Installation

1. Clone the repository:
```bash
git clone https://github.com/emilianoleite/telegram-bot.git
cd telegram-bot
```

2. Install dependencies:
```bash
go mod download
```

3. Build the bot:
```bash
go build -o telegram-bot
```

## Configuration

### Required Environment Variables

```bash
# Telegram Bot Configuration
export TELEGRAM_BOT_API_KEY="your_telegram_bot_token"

# LLM Provider (choose one)
export LLM_ADAPTER_TYPE="google_studio"  # or "huggingface"

# For Google AI Studio (recommended)
export GOOGLE_STUDIO_API_KEY="your_google_api_key"

# For Hugging Face
export HUGGINGFACEHUB_API_TOKEN="your_huggingface_token"
```

### Optional Environment Variables

```bash
# Output Destination (defaults to "telegram")
export OUTPUT_TYPE="telegram"  # or "google_sheets" or "file"

# For Google Sheets output
export GOOGLE_SHEET_ID="your_spreadsheet_id"
# Also place credentials.json in the project root

# For File output
export OUTPUT_PATH="./receipts"  # Directory to save receipt files
```

## Quick Start

### Basic Setup (Telegram Output)

The simplest configuration sends results back via Telegram:

```bash
export TELEGRAM_BOT_API_KEY="your_telegram_bot_token"
export GOOGLE_STUDIO_API_KEY="your_google_api_key"
export LLM_ADAPTER_TYPE="google_studio"
export OUTPUT_TYPE="telegram"

./telegram-bot
```

### Google Sheets Setup

To save receipts to Google Sheets:

1. Create a Google Cloud service account and download the credentials JSON
2. Share your Google Sheet with the service account email
3. Run the bot:

```bash
export TELEGRAM_BOT_API_KEY="your_telegram_bot_token"
export GOOGLE_STUDIO_API_KEY="your_google_api_key"
export LLM_ADAPTER_TYPE="google_studio"
export OUTPUT_TYPE="google_sheets"
export GOOGLE_SHEET_ID="your_spreadsheet_id"

./telegram-bot
```

### File Output Setup

To save receipts as local files:

```bash
export TELEGRAM_BOT_API_KEY="your_telegram_bot_token"
export GOOGLE_STUDIO_API_KEY="your_google_api_key"
export LLM_ADAPTER_TYPE="google_studio"
export OUTPUT_TYPE="file"
export OUTPUT_PATH="./receipts"

mkdir -p ./receipts
./telegram-bot
```

## Usage

1. Start the bot using one of the configurations above
2. Open Telegram and find your bot
3. Send a photo of a supermarket receipt
4. The bot will:
   - Download and process the image
   - Extract the list of items using AI
   - Send the results to your configured output destination

## Architecture

See [ADAPTER_USAGE.md](ADAPTER_USAGE.md) for detailed documentation on adapters.

This bot uses the **Adapter Pattern** to support multiple providers:

### LLM Adapters
- **Google AI Studio (Gemini)**: Best for receipt processing with multimodal support
- **Hugging Face**: Text-focused processing

### Output Adapters
- **Telegram**: Sends results as a message reply
- **Google Sheets**: Appends data to a spreadsheet
- **File**: Saves timestamped text files locally


## Project Structure

```
telegram-bot/
├── main.go                          # Main bot logic
├── adapters/
│   ├── llm_adapter.go               # LLM adapter interface
│   ├── output_adapter.go            # Output adapter interface
│   ├── factory.go                   # Factory functions for adapters
│   ├── huggingface_adapter.go       # Hugging Face implementation
│   ├── google_ai_studio/
│   │   └── main.go                  # Google AI Studio implementation
│   ├── telegram_output.go           # Telegram output implementation
│   ├── google_sheets.go             # Google Sheets output implementation
│   └── file_output.go               # File output implementation
├── go-huggingface/                  # Hugging Face client library
├── go.mod
├── go.sum
├── README.md
└── ADAPTER_USAGE.md                 # Detailed adapter documentation
```

## Development

### Adding a New LLM Provider

1. Create a new file in `adapters/` implementing the `LlmAdapter` interface
2. Add the adapter type to `llm_adapter.go`
3. Update the factory in `factory.go`

### Adding a New Output Destination

1. Create a new file in `adapters/` implementing the `OutputAdapter` interface
2. Add the output type to `output_adapter.go`
3. Update the factory in `factory.go`
