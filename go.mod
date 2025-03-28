module github.com/emilianoleite/telegram-bot

go 1.23.6

require (
	github.com/emilianoleite/telegram-bot/go-huggingface v0.0.15
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
)

replace github.com/emilianoleite/telegram-bot/go-huggingface => ./go-huggingface
