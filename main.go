package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	bot "example.com/hello_world_bot/bot"
	"github.com/joho/godotenv"
)

var botStatus int32 = 1  // 1 means online, 0 means offline

// Handler to show bot status
func statusHandler(w http.ResponseWriter, r *http.Request) {
    status := atomic.LoadInt32(&botStatus)
    if status == 1 {
        fmt.Fprintf(w, "<h1>Bot Status: Online</h1>")
    } else {
        fmt.Fprintf(w, "<h1>Bot Status: Offline</h1>")
    }
}

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
    go startBot()

    // Serve a simple status page on port 8099
    http.HandleFunc("/status", statusHandler)
    fmt.Println("Serving status on :8099")
    http.ListenAndServe(":8099", nil)
}
