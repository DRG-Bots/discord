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

	if len(content) < 3 {
		return
	}

	if content[:2] == "v " {
		rest := content[2:]
		var err error = nil
		switch rest {
		case "dd":
			message, err := buildOneDiveMessage(0)
			if err != nil {
				break
			}
			_, err = s.ChannelMessageSend(m.ChannelID, message)
		case "edd":
			message, err := buildOneDiveMessage(1)
			if err != nil {
				break
			}
			_, err = s.ChannelMessageSend(m.ChannelID, message)
		case "dives":
			message, err := buildBothDivesMessage()
			if err != nil {
				break
			}
			_, err = s.ChannelMessageSend(m.ChannelID, message)
		case "fact":
			_, err = s.ChannelMessageSend(m.ChannelID, getRandomLine(Trivia))
		default:
			return
		}
		if err != nil {
			fmt.Printf("Couldn't respond to message '%s' from user %s: %s\n", m.Content, m.Author.Username, err.Error())
		}
	}
}

func buildOneDiveMessage(diveId int) (string, error) {
	dives, err := getDivesData()
	if err != nil {
		return "", err
	}
	builder := strings.Builder{}
	builder.WriteString(buildDiveMessage(dives.Variants[diveId]))
	builder.WriteByte('\n')
	builder.WriteString(getRandomLine(Salutes))
	return builder.String(), nil
}

func buildBothDivesMessage() (string, error) {
	dives, err := getDivesData()
	if err != nil {
		return "", err
	}
	builder := strings.Builder{}
	builder.WriteString(buildDiveMessage(dives.Variants[0]))
	builder.WriteByte('\n')
	builder.WriteString(buildDiveMessage(dives.Variants[1]))
	builder.WriteByte('\n')
	builder.WriteString(getRandomLine(Salutes))
	builder.WriteString("\n\n")
	builder.WriteString("**Did You Know?** ")
	builder.WriteString(getRandomLine(Trivia))
	return builder.String(), nil
}

func buildDiveMessage(dive DeepDive) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("This week's **%s** is **%s**, set in the **%s**\n", dive.DDType, dive.Name, dive.Biome))
	for _, stage := range dive.Stages {
		warning := ""
		if stage.Warning != "" {
			warning = fmt.Sprintf(" warning: %s", stage.Warning)
		}
		anomaly := ""
		if stage.Anomaly != "" {
			anomaly = fmt.Sprintf(" Anomaly: %s", stage.Anomaly)
			if warning != "" {
				anomaly += ","
			}
		}
		sep := ""
		if anomaly != "" || warning != "" {
			sep = " |"
		}
		builder.WriteString(fmt.Sprintf("Stage **%d**: %s, %s%s%s%s\n", stage.Id, stage.Primary, stage.Secondary, sep, anomaly, warning))
	}
	return builder.String()
}

func getDivesData() (DeepDivesReqBody, error) {
	response, err := http.Get(DrgApiURL + "/v1/deepdives")
	if err != nil {
		return DeepDivesReqBody{}, err
	}
	defer response.Body.Close()
	if response.StatusCode == 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return DeepDivesReqBody{}, err
		}
		data := DeepDivesReqBody{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			return DeepDivesReqBody{}, err
		}
		return data, nil
	} else {
		return DeepDivesReqBody{}, errors.New(response.Status)
	}
}

func getRandomLine(lines []string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Intn(len(lines))
	return lines[i]
}
