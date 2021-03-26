package twitchchat

import (
	"bytes"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type MessageCommand int

const (
	UNKNOWN MessageCommand = iota
	CLEARCHAT
	CLEARMSG
	GLOBALUSERSTATE
	HOSTTARGET
	JOIN
	NOTICE
	PART
	PING
	PRIVMSG
	RECONNECT
	ROOMSTATE
	USERNOTICE
	USERSTATE
)

var MessageCommandLookup = map[string]MessageCommand{
	"CLEARCHAT":       CLEARCHAT,
	"CLEARMSG":        CLEARMSG,
	"GLOBALUSERSTATE": GLOBALUSERSTATE,
	"HOSTTARGET":      HOSTTARGET,
	"JOIN":            JOIN,
	"NOTICE":          NOTICE,
	"PART":            PART,
	"PING":            PING,
	"PRIVMSG":         PRIVMSG,
	"RECONNECT":       RECONNECT,
	"ROOMSTATE":       ROOMSTATE,
	"USERNOTICE":      USERNOTICE,
	"USERSTATE":       USERSTATE,
}

type ircPrefix struct {
	Nickname string
	User     string
	Host     string
}

type RawIrcMessage struct {
	RawMessage []byte
	RawTags    map[string]string
	ircPrefix
	RawCommand MessageCommand
	RawParams  [][]byte
}

type ClearChat struct {
	RawIrcMessage
	BanDuration uint
	Channel     string
	User        string
}

type ClearMsg struct {
	RawIrcMessage
	Channel     string
	Login       string
	Message     string
	TargetMsgId string
}

type GlobalUserState struct {
	RawIrcMessage
	BadgeInfo   string
	Badges      string
	Color       string
	DisplayName string
	EmoteSets   string
	Turbo       string
	UserId      string
	UserType    string
}

type HostTarget struct {
	RawIrcMessage
	Channel         string
	NumberOfViewers uint
	TargetChannel   string
}

type Join struct {
	RawIrcMessage
	Channel string
}

type Notice struct {
	RawIrcMessage
	Channel string
	Message string
	MsgId   string
}

type Part struct {
	RawIrcMessage
	Channel string
}

type Ping struct {
	RawIrcMessage
	Servers []string
}

type PrivMsg struct {
	RawIrcMessage
	BadgeInfo   string
	Badges      string
	Bits        string
	Channel     string
	Color       string
	DisplayName string
	Emotes      string
	Id          string
	Message     string
	Mod         string
	RoomId      string
	Subscriber  string
	TmiSentTs   string
	Turbo       string
	UserId      string
	UserType    string
}

type Reconnect struct {
	RawIrcMessage
}

type RoomState struct {
	RawIrcMessage
	Channel       string
	EmoteOnly     bool
	FollowersOnly int
	R9K           bool
	Slow          uint
	SubsOnly      bool
}

type UserNotice struct {
	RawIrcMessage
	BadgeInfo   string
	Badges      string
	Channel     string
	Color       string
	DisplayName string
	Emotes      string
	Id          string
	Login       string
	Message     string
	Mod         string
	MsgId       string
	RoomId      string
	Subscriber  bool
	SystemMsg   string
	TmiSentTs   string
	Turbo       string
	UserId      string
	UserType    string

	// On Sub.Resub
	// MsgParamCumulativeMonths string

	// // On Raid
	// MsgParamDisplayName string
	// MsgParamLogin       string

	// // On Subgift/anonsubgift
	// MsgParamMonths string

	// // anongiftpaidupgrade/giftpaidupgrade
	// MsgParamPromoGiftTotal       string
	// MsgParamPromoName            string
	// MsgParamRecipientDisplayName string
	// MsgParamRecipientId          string
	// MsgParamRecipientUserName    string
	// MsgParamSenderLogin          string
	// MsgParamSenderName           string
	// MsgParamShouldShareStreak    string
	// MsgParamStreakMonths         string
	// MsgParamSubPlan              string
	// MsgParamSubPlanName          string
	// MsgParamViewerCount          string
	// MsgParamRitualName           string
	// MsgParamThreshold            string
	// MsgParamGiftMonths           string
}

type UserState struct {
	RawIrcMessage
	BadgeInfo   string
	Badges      string
	Channel     string
	Color       string
	DisplayName string
	EmoteSets   string
	Mod         string
	Subscriber  bool
	Turbo       string
	UserType    string
}

func getStringFromTags(tags map[string]string, key string) string {
	if str, ok := tags[key]; ok {
		return str
	} else {
		return ""
	}
}

func getChannel(params [][]byte) string {
	if len(params) > 0 {
		if bytes.HasPrefix(params[0], []byte("#")) {
			return string(params[0][1:])
		}
	}
	return ""
}

func getUintFromTags(tags map[string]string, key string) uint {
	str := getStringFromTags(tags, key)
	if str != "" {
		if num, err := strconv.ParseUint(str, 10, 32); err == nil {
			return uint(num)
		}
	}
	return 0
}

func getIntFromTags(tags map[string]string, key string) int {
	str := getStringFromTags(tags, key)
	if str != "" {
		if num, err := strconv.ParseInt(str, 10, 32); err == nil {
			return int(num)
		}
	}
	return 0
}

func getBoolFromTags(tags map[string]string, key string) bool {
	str := getStringFromTags(tags, key)
	if str != "" {
		if num, err := strconv.ParseUint(str, 10, 32); err == nil {
			// 1 is true, 0 is false
			return num == 1
		}
	}
	return false
}

func newClearChatMsg(rawMsg RawIrcMessage) *ClearChat {
	msg := ClearChat{
		RawIrcMessage: rawMsg,
		BanDuration:   getUintFromTags(rawMsg.RawTags, "ban-duration"),
		Channel:       getChannel(rawMsg.RawParams),
	}

	if len(rawMsg.RawParams) > 1 {
		if bytes.HasPrefix(rawMsg.RawParams[1], []byte(":")) {
			msg.User = string(rawMsg.RawParams[1][1:])
		}
	}

	return &msg
}

func newClearMsgMsg(rawMsg RawIrcMessage) *ClearMsg {
	msg := ClearMsg{
		RawIrcMessage: rawMsg,
		Channel:       getChannel(rawMsg.RawParams),
		Login:         getStringFromTags(rawMsg.RawTags, "login"),
		TargetMsgId:   getStringFromTags(rawMsg.RawTags, "target-msg-id"),
	}

	// Params[0] should be the Channel, the rest are the message
	if len(msg.RawParams) > 1 {
		msg.Message = string(bytes.Join(msg.RawParams[1:], []byte(" ")))
		msg.Message = strings.TrimPrefix(msg.Message, ":")
	}

	return &msg
}

func newGlobalUserStateMsg(rawMsg RawIrcMessage) *GlobalUserState {
	msg := GlobalUserState{
		RawIrcMessage: rawMsg,
		BadgeInfo:     getStringFromTags(rawMsg.RawTags, "badge-info"),
		Badges:        getStringFromTags(rawMsg.RawTags, "badges"),
		Color:         getStringFromTags(rawMsg.RawTags, "color"),
		DisplayName:   getStringFromTags(rawMsg.RawTags, "display-name"),
		EmoteSets:     getStringFromTags(rawMsg.RawTags, "emote-sets"),
		Turbo:         getStringFromTags(rawMsg.RawTags, "turbo"),
		UserId:        getStringFromTags(rawMsg.RawTags, "user-id"),
		UserType:      getStringFromTags(rawMsg.RawTags, "user-type"),
	}

	return &msg
}

func newHostTargetMsg(rawMsg RawIrcMessage) *HostTarget {
	msg := HostTarget{
		RawIrcMessage: rawMsg,
		Channel:       getChannel(rawMsg.RawParams),
	}

	if len(rawMsg.RawParams) > 1 {
		msg.TargetChannel = string(bytes.TrimPrefix(rawMsg.RawParams[1], []byte(":")))
	}

	if len(rawMsg.RawParams) > 2 {
		if num, err := strconv.ParseUint(string(rawMsg.RawParams[2]), 10, 32); err == nil {
			msg.NumberOfViewers = uint(num)
		}
	}

	return &msg
}

func newJoinMsg(rawMsg RawIrcMessage) *Join {
	msg := Join{
		RawIrcMessage: rawMsg,
		Channel:       getChannel(rawMsg.RawParams),
	}

	return &msg
}

func newNoticeMsg(rawMsg RawIrcMessage) *Notice {
	msg := Notice{
		RawIrcMessage: rawMsg,
		Channel:       getChannel(rawMsg.RawParams),
		MsgId:         getStringFromTags(rawMsg.RawTags, "msg-id"),
	}

	// Params[0] should be the Channel, the rest are the message
	if len(msg.RawParams) > 1 {
		msg.Message = string(bytes.Join(msg.RawParams[1:], []byte(" ")))
		msg.Message = strings.TrimPrefix(msg.Message, ":")
	}

	return &msg
}

func newPartMsg(rawMsg RawIrcMessage) *Part {
	msg := Part{
		RawIrcMessage: rawMsg,
		Channel:       getChannel(rawMsg.RawParams),
	}

	return &msg
}

func newPingMsg(rawMsg RawIrcMessage) *Ping {
	msg := Ping{
		RawIrcMessage: rawMsg,
	}

	msg.Servers = make([]string, len(rawMsg.RawParams))
	for i, server := range rawMsg.RawParams {
		msg.Servers[i] = string(server)
	}

	return &msg
}

func newPrivMsgMsg(rawMsg RawIrcMessage) *PrivMsg {
	msg := PrivMsg{
		RawIrcMessage: rawMsg,
		BadgeInfo:     getStringFromTags(rawMsg.RawTags, "badge-info"),
		Badges:        getStringFromTags(rawMsg.RawTags, "badges"),
		Bits:          getStringFromTags(rawMsg.RawTags, "bits"),
		Channel:       getChannel(rawMsg.RawParams),
		Color:         getStringFromTags(rawMsg.RawTags, "color"),
		DisplayName:   getStringFromTags(rawMsg.RawTags, "display-name"),
		Emotes:        getStringFromTags(rawMsg.RawTags, "emotes"),
		Id:            getStringFromTags(rawMsg.RawTags, "id"),
		Mod:           getStringFromTags(rawMsg.RawTags, "mod"),
		RoomId:        getStringFromTags(rawMsg.RawTags, "room-id"),
		Subscriber:    getStringFromTags(rawMsg.RawTags, "subscriber"),
		TmiSentTs:     getStringFromTags(rawMsg.RawTags, "tmi-sent-ts"),
		Turbo:         getStringFromTags(rawMsg.RawTags, "turbo"),
		UserId:        getStringFromTags(rawMsg.RawTags, "user-id"),
		UserType:      getStringFromTags(rawMsg.RawTags, "user-type"),
	}

	// Params[0] should be the Channel, the rest are the message
	if len(msg.RawParams) > 1 {
		msg.Message = string(bytes.Join(msg.RawParams[1:], []byte(" ")))
		msg.Message = strings.TrimPrefix(msg.Message, ":")
	}

	return &msg
}

func newReconnectMsg(rawMsg RawIrcMessage) *Reconnect {
	msg := Reconnect{
		RawIrcMessage: rawMsg,
	}

	return &msg
}

func newRoomStateMsg(rawMsg RawIrcMessage) *RoomState {
	msg := RoomState{
		RawIrcMessage: rawMsg,
		EmoteOnly:     getBoolFromTags(rawMsg.RawTags, "emote-only"),
		FollowersOnly: getIntFromTags(rawMsg.RawTags, "emote-only"),
		R9K:           getBoolFromTags(rawMsg.RawTags, "r9k"),
		Slow:          getUintFromTags(rawMsg.RawTags, "slow"),
		SubsOnly:      getBoolFromTags(rawMsg.RawTags, "subs-only"),
	}

	return &msg
}

func newUserNoticeMsg(rawMsg RawIrcMessage) *UserNotice {
	msg := UserNotice{
		RawIrcMessage: rawMsg,
		BadgeInfo:     getStringFromTags(rawMsg.RawTags, "badge-info"),
		Badges:        getStringFromTags(rawMsg.RawTags, "badges"),
		Channel:       getChannel(rawMsg.RawParams),
		Color:         getStringFromTags(rawMsg.RawTags, "color"),
		DisplayName:   getStringFromTags(rawMsg.RawTags, "display-name"),
		Emotes:        getStringFromTags(rawMsg.RawTags, "emotes"),
		Id:            getStringFromTags(rawMsg.RawTags, "id"),
		Login:         getStringFromTags(rawMsg.RawTags, "login"),
		Mod:           getStringFromTags(rawMsg.RawTags, "mod"),
		MsgId:         getStringFromTags(rawMsg.RawTags, "msg-id"),
		RoomId:        getStringFromTags(rawMsg.RawTags, "room-id"),
		Subscriber:    getBoolFromTags(rawMsg.RawTags, "subscriber"),
		TmiSentTs:     getStringFromTags(rawMsg.RawTags, "tmi-sent-ts"),
		Turbo:         getStringFromTags(rawMsg.RawTags, "turbo"),
		UserId:        getStringFromTags(rawMsg.RawTags, "user-id"),
		UserType:      getStringFromTags(rawMsg.RawTags, "user-type"),
	}

	// Params[0] should be the Channel, the rest are the message
	if len(msg.RawParams) > 1 {
		msg.Message = string(bytes.Join(msg.RawParams[1:], []byte(" ")))
		msg.Message = strings.TrimPrefix(msg.Message, ":")
	}

	return &msg
}

func newUserStateMsg(rawMsg RawIrcMessage) *UserState {
	msg := UserState{
		RawIrcMessage: rawMsg,
		BadgeInfo:     getStringFromTags(rawMsg.RawTags, "badge-info"),
		Badges:        getStringFromTags(rawMsg.RawTags, "badges"),
		Channel:       getChannel(rawMsg.RawParams),
		Color:         getStringFromTags(rawMsg.RawTags, "color"),
		DisplayName:   getStringFromTags(rawMsg.RawTags, "display-name"),
		EmoteSets:     getStringFromTags(rawMsg.RawTags, "emote-sets"),
		Mod:           getStringFromTags(rawMsg.RawTags, "mod"),
		Subscriber:    getBoolFromTags(rawMsg.RawTags, "subscriber"),
		Turbo:         getStringFromTags(rawMsg.RawTags, "turbo"),
		UserType:      getStringFromTags(rawMsg.RawTags, "user-type"),
	}

	return &msg
}

// Empty interface for handling IRC messages
type IrcMessage interface{}

func newIrcMessage(rawMsg RawIrcMessage) IrcMessage {
	var rval IrcMessage

	rval = &rawMsg

	switch rawMsg.RawCommand {
	case CLEARCHAT:
		rval = newClearChatMsg(rawMsg)
	case CLEARMSG:
		rval = newClearMsgMsg(rawMsg)
	case GLOBALUSERSTATE:
		rval = newGlobalUserStateMsg(rawMsg)
	case HOSTTARGET:
		rval = newHostTargetMsg(rawMsg)
	case JOIN:
		rval = newJoinMsg(rawMsg)
	case NOTICE:
		rval = newNoticeMsg(rawMsg)
	case PART:
		rval = newPartMsg(rawMsg)
	case PING:
		rval = newPingMsg(rawMsg)
	case PRIVMSG:
		rval = newPrivMsgMsg(rawMsg)
	case RECONNECT:
		rval = newReconnectMsg(rawMsg)
	case ROOMSTATE:
		rval = newRoomStateMsg(rawMsg)
	case USERNOTICE:
		rval = newUserNoticeMsg(rawMsg)
	case USERSTATE:
		rval = newUserStateMsg(rawMsg)
	}

	return rval
}

func bytesToIrcMessage(buffer []byte) IrcMessage {

	rawMsg := RawIrcMessage{
		RawMessage: buffer,
		RawCommand: UNKNOWN,
		RawTags:    make(map[string]string),
	}

	msgPieces := bytes.Split(buffer, []byte(" "))
	currentIndex := 0

	// Tags are included
	// @<key>=<val>;<key>=<val>....
	if bytes.HasPrefix(msgPieces[currentIndex], []byte("@")) {
		msgPieces[currentIndex] = bytes.TrimPrefix(msgPieces[currentIndex], []byte("@"))
		tags := bytes.Split(msgPieces[currentIndex], []byte(";"))
		for i := range tags {
			keyVal := bytes.Split(tags[i], []byte("="))
			if len(keyVal) == 2 {
				rawMsg.RawTags[string(keyVal[0])] = string(keyVal[1])
			}
		}
		currentIndex++
	}

	// Message contains source information
	// :tmi.twitch.tv or :<user>!<user>@<user>.tmi.twitch.tv
	if bytes.HasPrefix(msgPieces[currentIndex], []byte(":")) {
		msgPieces[currentIndex] = bytes.TrimPrefix(msgPieces[currentIndex], []byte(":"))

		pieces := regexp.MustCompile(`!|@`).Split(string(msgPieces[currentIndex]), -1)

		switch len(pieces) {
		case 1:
			rawMsg.Host = pieces[0]
		case 3:
			rawMsg.Nickname = pieces[0]
			rawMsg.User = pieces[1]
			rawMsg.Host = pieces[2]
		default:
			log.Println(pieces)
		}
		currentIndex++
	}

	if len(msgPieces) > currentIndex {
		rawMsg.RawCommand = MessageCommandLookup[string(msgPieces[currentIndex])]
		currentIndex++
	}

	if len(msgPieces) > currentIndex {
		rawMsg.RawParams = msgPieces[currentIndex:]
	}

	return newIrcMessage(rawMsg)
}
