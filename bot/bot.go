package Bot

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var BotToken string

var Command = struct {
	ABOUT string
	HELP  string
	ROLL  string
	DAYS  string
}{
	ABOUT: "about",
	HELP:  "help",
	ROLL:  "roll",
	DAYS:  "days",
}

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
			channelID := "1253715243474096238"

			handleDays(discord, channelID, "3", "7")

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

func handlePrintingDays(discord *discordgo.Session, channelId string, startDate time.Time, daysCount int) {
	// Initial message before listing days
	initialMessage := "@everyone Play when? (add a 🐲 if you wanna play or a 🧙‍♂️ if you wanna DM)"
	discord.ChannelMessageSend(channelId, initialMessage)

	// Loop over the next X days starting from the calculated startDate
	for i := 0; i < daysCount; i++ {
		// Calculate the future date (startDate + i days)
		day := startDate.AddDate(0, 0, i)

		// Get the day number and suffix
		dayNum := day.Day()
		suffix := getDaySuffix(dayNum)

		// Format the day as "Monday 20th September"
		formattedDay := fmt.Sprintf("%s %d%s %s", day.Format("Monday"), dayNum, suffix, day.Format("January"))

		// Send each day as a separate message
		discord.ChannelMessageSend(channelId, formattedDay)

		// Add a delay to avoid rate-limiting issues
		time.Sleep(2 * time.Second) // Adjust the delay as needed
	}

	// Final message
	finalMessage := "If you see the above but can't play these dates, please kindly respond to this message with an emoji"
	discord.ChannelMessageSend(channelId, finalMessage)
}

func handleDays(discord *discordgo.Session, channelId string, daysFromToday string, daysCount string) {
	start, err1 := strconv.Atoi(daysFromToday)
	count, err2 := strconv.Atoi(daysCount)

	if daysFromToday == "" || err1 != nil || start < 0 {
		discord.ChannelMessageSend(channelId, "I need a number for how many days from now should i start counting, for example 'days <1> 7'")
		return
	}
	if daysCount == "" || err2 != nil || count < 1 {
		discord.ChannelMessageSend(channelId, "I need a number for how many days to count to after the start date, for example 'days 1 <7>'")
		return
	}

	if count > 14 {
		discord.ChannelMessageSend(channelId, "Printing more than 2 weeks seems a bit excessive, don't ya think?")
		return
	}

	currentTime := time.Now()
	startDate := currentTime.AddDate(0, 0, start)

	handlePrintingDays(discord, channelId, startDate, count)
}

func handleAbout(discord *discordgo.Session, message *discordgo.MessageCreate) {
	discord.ChannelMessageSend(message.ChannelID, "I'm a bot that helps printing scheduling messages and roll character stats, coded in Go by Javier Baccarelli. You can checkout my code over here https://github.com/Geddard/dnd-goblin")
}

func handleHelp(discord *discordgo.Session, message *discordgo.MessageCreate) {
	discord.ChannelMessageSend(message.ChannelID, `This is how to use me:
- days: Print a scheduling message by sending "days <x> <y>" where <x> is replaced by a number of days from today (0 for today, 1 for tomorrow, and so on) and <y> should be replaced by the number of days to print after the start date, minimum 1.
- roll: Roll basic D&D Character Stats.
- about: Basic info about me.
- help: Print this same message.`)
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

	content := message.Content

	switch {
	case strings.Contains(content, Command.ABOUT):
		handleAbout(discord, message)
	case strings.Contains(content, Command.HELP):
		handleHelp(discord, message)
	case strings.Contains(content, Command.ROLL):
		handleRoll(discord, message)
	case strings.Contains(content, Command.DAYS):
		// Find the index of the "days" command
		index := strings.Index(content, Command.DAYS)
		if index == -1 {
			return // If "days" is not found, exit
		}

		// Extract parameters after the "days" command
		params := strings.Fields(content[index:])

		// Ensure there are at least two parameters after "days"
		var param1, param2 string
		if len(params) >= 2 {
			param1 = params[1]
		} else {
			param1 = ""
			return
		}
		if len(params) >= 3 {
			param2 = params[2]
		} else {
			param2 = ""
		}

		handleDays(discord, message.ChannelID, param1, param2)
	}
}
