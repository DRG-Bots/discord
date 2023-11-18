package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Token   string
	Salutes []string
	Trivia  []string
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
	// Get salutes and trivia from API
	var err error // needed to not redeclare variables
	Salutes, err = getApiStringList("/v1/salutes", "salutes")
	if err != nil {
		fmt.Println(err)
		return
	}
	Trivia, err = getApiStringList("/v1/trivia", "trivia")
	if err != nil {
		fmt.Println(err)
		return
	}

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

func getApiStringList(endpoint, key string) ([]string, error) {
	response, err := http.Get(DrgApiURL + endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		var data map[string][]string
		err = json.Unmarshal(body, &data)
		if err != nil {
			return nil, err
		}
		return data[key], nil
	} else {
		return nil, errors.New("Endpoint " + endpoint + " returned status code " + response.Status)
	}
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
		_, err := s.ChannelMessageSend(m.ChannelID, getRandomLine(Salutes))
		if err != nil {
			fmt.Printf("Couldn't respond to message '%s' from user %s", m.Content, m.Author.Username)
		}
		return
	}
}

func getRandomLine(lines []string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Intn(len(lines))
	return lines[i]
}
