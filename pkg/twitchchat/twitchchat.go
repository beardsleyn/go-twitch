package twitchchat

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const twitchChatUrl = "ws://irc-ws.chat.twitch.tv:80"

type chatMsg struct {
	channel string
	message string
}

type chatEmitter struct {
	Emitter
	irc *Irc
}

func newChatEmitter(irc *Irc) *chatEmitter {
	return &chatEmitter{
		irc: irc,
	}
}

func (em *chatEmitter) Emit(event Event) error {
	msg, ok := event.(chatMsg)
	if !ok {
		// todo
		return nil
	}
	return em.irc.Privmsg(msg.channel, msg.message)
}

type joinEmitter struct {
	Emitter
	irc *Irc
}

func newJoinEmitter(irc *Irc) *joinEmitter {
	return &joinEmitter{
		irc: irc,
	}
}

func (em *joinEmitter) Emit(event Event) error {
	channel, ok := event.(string)
	if !ok {
		// todo
		return nil
	}

	return em.irc.Join(channel)
}

type Options struct {
	Nick       string
	Pass       string
	ChatLimit  int // Defaults to 20
	JoinLimit  int // Defaults to 20
	AuthLimit  int // Defaults to 20
	EnableTags bool
}

type TwitchChat struct {
	irc           *Irc
	ircChan       chan IrcMessage
	options       Options
	messageRouter map[string]interface{}
	privMsgBucket *Bucket
	joinBucket    *Bucket

	joinChannelMutex sync.RWMutex
	joinedChannels   map[string]bool
}

func NewTwitchChat(options *Options) (*TwitchChat, error) {

	if options.Nick == "" {
		return nil, errors.New("no nick provided")
	}
	if options.Pass == "" {
		return nil, errors.New("no pass provided")
	}

	tc := new(TwitchChat)
	tc.options = *options

	if tc.options.ChatLimit == 0 {
		tc.options.ChatLimit = 20
	}
	if tc.options.JoinLimit == 0 {
		tc.options.JoinLimit = 20
	}
	if tc.options.AuthLimit == 0 {
		tc.options.AuthLimit = 20
	}

	tc.messageRouter = make(map[string]interface{})

	tc.joinedChannels = make(map[string]bool)

	var err error
	tc.irc, err = NewIrc()

	tc.RegisterCallback(tc.Pong)

	tc.privMsgBucket = NewBucket(newChatEmitter(tc.irc),
		rate.Every(time.Duration(30/tc.options.ChatLimit)*time.Second), 1)
	tc.joinBucket = NewBucket(newJoinEmitter(tc.irc),
		rate.Every(time.Duration(30/tc.options.JoinLimit)*time.Second), 1)
	return tc, err
}

func (tc *TwitchChat) Connect() error {
	tc.ircChan = make(chan IrcMessage)

	go tc.handleIrcMessage()

	return tc.irc.Connect(tc.options.Nick, tc.options.Pass, tc.options.EnableTags, tc.ircChan)
}

func (tc *TwitchChat) Disconnect() error {
	err := tc.irc.Disconnect()
	tc.joinedChannels = make(map[string]bool)
	return err
}

func (tc *TwitchChat) handleIrcMessage() {
	for msg := range tc.ircChan {
		arg := reflect.ValueOf(msg)
		if cb, ok := tc.messageRouter[arg.Type().Elem().String()]; ok {
			refCb := reflect.ValueOf(cb)
			if refCb.Kind() == reflect.Func {
				refCb.Call([]reflect.Value{arg})
			}
		}
	}
}

func (tc *TwitchChat) RegisterCallback(cb interface{}) error {
	v := reflect.ValueOf(cb)
	if v.Kind() != reflect.Func {
		return errors.New("not a function")
	}

	if v.Type().NumIn() != 1 {
		return errors.New("too many args")
	}

	ti := v.Type().In(0)
	if ti.Kind() != reflect.Ptr {
		return errors.New("wrong type of arg")
	}

	tc.messageRouter[ti.Elem().String()] = cb

	return nil
}

func (tc *TwitchChat) Chat(channel, msg string) error {
	tc.privMsgBucket.AddEvent(chatMsg{
		channel: channel,
		message: msg,
	}, false)
	return nil
}

func (tc *TwitchChat) Join(channel string) error {
	defer tc.joinChannelMutex.Unlock()
	tc.joinBucket.AddEvent(channel, false)
	tc.joinChannelMutex.Lock()
	tc.joinedChannels[channel] = true
	return nil
}

func (tc *TwitchChat) Part(channel string) error {
	defer tc.joinChannelMutex.Unlock()
	tc.joinChannelMutex.Lock()
	delete(tc.joinedChannels, channel)
	return tc.irc.Part(channel)
}

func (tc *TwitchChat) Pong(ping *Ping) {
	for _, server := range ping.Servers {
		err := tc.irc.Pong(server)
		if err != nil {
			// todo
		}
	}
}
