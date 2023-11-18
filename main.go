package main

import (
	// "encoding/json"
	"fmt"
	// "io"
	// "net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	Token string
)

const (
	DrgApiURL = "https://drgapi.com"
	TokenFile = "token.txt"
)

func init() {
	f, err := os.Open(TokenFile)
	if err != nil {
		panic("token.txt file not found!")
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		panic("Could not read token.txt")
	}
	Token = string(buf[:n])
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

type Gopher struct {
	Name string `json:"name"`
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	content := strings.ToLower(m.Content)

	if content == "v" || (strings.Contains(content, "rock") && strings.Contains(content, "stone")) {
		_, err := s.ChannelMessageSend(m.ChannelID, "Rock and Stone!")
		if err != nil {
			fmt.Printf("Couldn't respond to message '%s' from user %s", m.Content, m.Author.Username)
		}
	}
}
