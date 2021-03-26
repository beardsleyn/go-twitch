package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/beardsleyn/go-twitch/pkg/twitchchat"
)

type client struct {
	Tc *twitchchat.TwitchChat
}

func (c client) OnPriv(priv *twitchchat.PrivMsg) {
	log.Println(priv.User + ": " + priv.Message)
	if strings.HasPrefix(priv.Message, "!8ball") {
		c.Tc.Chat(priv.Channel, "It is decidedly so.")
	}
}

func (c client) OnNotice(not *twitchchat.Notice) {
	log.Println(not.Message)
}

func (c client) OnPart(part *twitchchat.Part) {
	log.Println("Parted: " + part.Nickname)
}

func (c client) OnJoin(join *twitchchat.Join) {
	log.Println("Joined: " + join.User)
}

// A message that was not parsed properly. Generally the welcome messages
func (c client) OnRawMsg(raw *twitchchat.RawIrcMessage) {
	log.Println(string(raw.RawMessage))
}

func NewClient(opts *twitchchat.Options) (*client, error) {

	tc, err := twitchchat.NewTwitchChat(opts)
	if err != nil {
		return nil, err
	}

	client := new(client)
	client.Tc = tc

	tc.RegisterCallback(client.OnPriv)
	tc.RegisterCallback(client.OnNotice)
	tc.RegisterCallback(client.OnPart)
	tc.RegisterCallback(client.OnJoin)
	tc.RegisterCallback(client.OnRawMsg)

	return client, nil
}

func main() {

	client, err := NewClient(&twitchchat.Options{
		Nick:       os.Getenv("NICK"),
		Pass:       os.Getenv("PASS"),
		EnableTags: true,
	})

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	client.Tc.Connect()

	client.Tc.Join("deovontay_mcslanga")

	reader := bufio.NewReader(os.Stdin)
	for {
		// read line from console
		msg, msgErr := reader.ReadString('\n')
		if msgErr != nil {
			break
		}

		client.Tc.Chat("deovontay_mcslanga", msg)
	}

	client.Tc.Disconnect()

	os.Exit(0)
}
