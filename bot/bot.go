package Bot

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var BotToken string

func checkNilErr(e error) {
	if e != nil {
		log.Fatal("Error message")
	}
}

func Run() {
	// create a session
	discord, err := discordgo.New("Bot " + BotToken)
	checkNilErr(err)

	// add an event handler
	discord.AddHandler(messages)

	// open session
	discord.Open()
	defer discord.Close() // close session, after function termination

	// Start a goroutine to check time and send `!days` every Friday at 6pm
	go scheduleDaysMessage(discord)

	// keep bot running until there is NO os interruption (ctrl + C)
	fmt.Println("Bot running....")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func scheduleDaysMessage(discord *discordgo.Session) {
	for {
		now := time.Now()

		// Check if today is Friday and it's 6pm
		if now.Weekday() == time.Friday && now.Hour() == 18 && now.Minute() == 0 {
			// Replace "YOUR_CHANNEL_ID" with the actual channel ID you want to send the message to
			channelID := "1213948847471067146"

			// Simulate sending the `!days` command automatically
			handleDays(discord, &discordgo.MessageCreate{Message: &discordgo.Message{ChannelID: channelID}}, 3)

			// Sleep for a full minute to avoid sending multiple messages within the same minute
			time.Sleep(time.Minute)
		}

		// Sleep for 30 seconds before checking again
		time.Sleep(30 * time.Second)
	}
}

func getDaySuffix(day int) string {
	switch day {
	case 1, 21, 31:
		return "st"
	case 2, 22:
		return "nd"
	case 3, 23:
		return "rd"
	default:
		return "th"
	}
}

func handleDays(discord *discordgo.Session, message *discordgo.MessageCreate, days int) {
	// Get the current date
	currentTime := time.Now()

	// Calculate the start date by adding 'days' to the current date
	startDate := currentTime.AddDate(0, 0, days)

	// Initial message before listing days
	initialMessage := "@everyone Play when? (add a 🐲 if you wanna play or a 🧙‍♂️ if you wanna DM)"
	discord.ChannelMessageSend(message.ChannelID, initialMessage)

	// Loop over the next 7 days starting from the calculated startDate
	for i := 0; i < 7; i++ {
		// Calculate the future date (startDate + i days)
		day := startDate.AddDate(0, 0, i)

		// Get the day number and suffix
		dayNum := day.Day()
		suffix := getDaySuffix(dayNum)

		// Format the day as "Monday 20th September"
		formattedDay := fmt.Sprintf("%s %d%s %s", day.Format("Monday"), dayNum, suffix, day.Format("January"))

		// Send each day as a separate message
		discord.ChannelMessageSend(message.ChannelID, formattedDay)

		// Add a delay to avoid rate-limiting issues
		time.Sleep(2 * time.Second) // Adjust the delay as needed
	}

	// Final message
	finalMessage := "If you see the above but can't play this week, please kindly respond to this message with an emoji"
	discord.ChannelMessageSend(message.ChannelID, finalMessage)
}

func handleAbout(discord *discordgo.Session, message *discordgo.MessageCreate) {
	discord.ChannelMessageSend(message.ChannelID, "I'm a bot that helps printing scheduling messages and roll character stats, coded in Go by Javier Baccarelli")
}

func rollDice() int {
	return rand.Intn(6) + 1
}

func roll4d6() int {
	dice := make([]int, 4)
	for i := 0; i < 4; i++ {
		dice[i] = rollDice()
	}
	sort.Ints(dice)                    // Sort dice to drop the lowest
	return dice[1] + dice[2] + dice[3] // Sum the top 3
}

func handleRoll(discord *discordgo.Session, message *discordgo.MessageCreate) {
	rand.New(rand.NewSource(time.Now().UnixNano())) // Seed random number generator
	var stats [6]int
	for i := 0; i < 6; i++ {
		stats[i] = roll4d6()
	}

	statStrings := make([]string, len(stats))
	for i, stat := range stats {
		statStrings[i] = fmt.Sprintf("%d", stat)
	}
	// Join the slice with commas
	statsMessage := fmt.Sprintf("Your D&D character stats are: %s", strings.Join(statStrings, ", "))

	discord.ChannelMessageSend(message.ChannelID, statsMessage)
}

func messages(discord *discordgo.Session, message *discordgo.MessageCreate) {
	// prevent bot responding to its own message
	if message.Author.ID == discord.State.User.ID {
		return
	}

	switch {
	case strings.Contains(message.Content, "!about"):
		handleAbout(discord, message)
	case strings.Contains(message.Content, "!roll"):
		handleRoll(discord, message)
	case strings.Contains(message.Content, "!days"):
		handleDays(discord, message, 1)
	}
}
