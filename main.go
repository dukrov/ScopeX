package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Config struct {
	Token           string   `json:"token"`
	AuthorizedUsers []int64  `json:"authorized_users"`
}

var config Config

func main() {
	// Load config
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal("Error reading config:", err)
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&config)

	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Authorized on account", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		userID := update.Message.From.ID
		msgText := update.Message.Text

		logCommand(userID, msgText)

		if !isAuthorized(userID) {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unauthorized access."))
			continue
		}

		var output string
		switch strings.ToLower(msgText) {
		case "info serv":
			output = runCommand("uname", "-a")
		case "ip route":
			output = runCommand("ip", "route")
		case "public ip":
			output = runCommand("curl", "-s", "ifconfig.me")
		case "active processes":
			output = runCommand("ps", "aux")
		case "net":
			output = runCommand("netstat", "-tulnp")
		default:
			output = "Unknown command."
		}

		// Truncate long messages
		if len(output) > 4000 {
			output = output[:3999] + "…"
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "```\n"+output+"\n```")
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	}
}

func isAuthorized(userID int) bool {
	for _, id := range config.AuthorizedUsers {
		if id == int64(userID) {
			return true
		}
	}
	return false
}

func runCommand(name string, args ...string) string {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return "Error: " + err.Error()
	}
	return string(out)
}

func logCommand(userID int, command string) {
	f, err := os.OpenFile("logFileBot.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		logEntry := strings.Join([]string{time.Now().Format(time.RFC3339), "UserID:", string(rune(userID)), "Command:", command, "\n"}, " ")
		f.WriteString(logEntry)
	}
}
