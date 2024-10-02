package main

import (
	"log"
	"os"

	bot "example.com/hello_world_bot/bot"
	"github.com/joho/godotenv"
)

func startBot() {
	// Load environment variables from .env file (if applicable)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found; using environment variables.")
	}

	// Get the bot token from the environment variable
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		log.Fatalf("Bot token not found in environment variables")
	}

	// Assign the bot token to the bot's BotToken variable
	bot.BotToken = botToken

	// Call the run function of bot/bot.go
	bot.Run()
}

func main() {
	// Start the bot in a separate goroutine
	startBot()
}
