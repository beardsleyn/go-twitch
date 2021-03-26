package twitchchat

import (
	"bytes"
	"log"

	"github.com/gorilla/websocket"
)

type Irc struct {
	ws      *websocket.Conn
	OutChan chan<- IrcMessage
	rcvChan chan []byte
}

func (irc *Irc) Connect(user string, pass string, tags bool, outChan chan<- IrcMessage) error {
	sock, _, err := websocket.DefaultDialer.Dial(twitchChatUrl, nil)
	if err != nil {
		return err
	}
	irc.ws = sock
	irc.OutChan = outChan
	irc.rcvChan = make(chan []byte)

	go irc.handleReceivedMessage()

	go func() {
		defer close(irc.rcvChan)
		for {
			_, message, err := irc.ws.ReadMessage()
			if err != nil {
				return
			}
			irc.rcvChan <- message
		}
	}()

	err = irc.ws.WriteMessage(websocket.TextMessage, []byte("CAP REQ :twitch.tv/tags twitch.tv/commands twitch.tv/membership"))
	if err != nil {
		return err
	}

	err = irc.sendBytes([]byte("PASS oauth:" + pass))
	if err != nil {
		log.Println("Couldn't write PASS")
		return err
	}
	err = irc.sendBytes([]byte("NICK " + user))
	if err != nil {
		log.Println("Couldn't write NICK")
		return err
	}

	return nil
}

func (irc *Irc) Disconnect() error {
	return irc.ws.Close()
}

func (irc *Irc) handleReceivedMessage() {
	defer close(irc.OutChan)
	for rcvMsg := range irc.rcvChan {
		lines := bytes.Split(rcvMsg, []byte("\r\n"))
		for _, msgBytes := range lines {
			if len(msgBytes) > 0 {
				ircMsg := bytesToIrcMessage(msgBytes)
				irc.OutChan <- ircMsg
			}
		}
	}
}

func (irc *Irc) sendBytes(bytes []byte) error {
	err := irc.ws.WriteMessage(websocket.TextMessage, bytes)
	return err
}

func (irc *Irc) Join(channel string) error {
	return irc.sendBytes([]byte("JOIN #" + channel + "\r\n"))
}

func (irc *Irc) Part(channel string) error {
	return irc.sendBytes([]byte("PART #" + channel + "\r\n"))
}

func (irc *Irc) Pong(server string) error {
	return irc.sendBytes([]byte("PONG " + server + "\r\n"))
}

func (irc *Irc) Privmsg(channel, msg string) error {
	return irc.sendBytes([]byte("PRIVMSG #" + channel + " :" + msg + "\r\n"))
}

func (irc *Irc) CapReq(req string) error {
	return nil
}

func NewIrc() (*Irc, error) {

	irc := new(Irc)

	return irc, nil
}
