package twitchchat

import (
	"fmt"
	"testing"
)

func Test_bytesToIrcMessage(t *testing.T) {
	var bytes []byte
	var ircMsg IrcMessage

	// "CLEARCHAT":       CLEARCHAT,
	bytes = []byte(":tmi.twitch.tv CLEARCHAT #dallas")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*ClearChat); ok {
		if msg.Channel != "dallas" {
			t.Error("Wrong Channel: " + msg.Channel)
		}
		if msg.Host != "tmi.twitch.tv" {
			t.Error("Wrong host")
		}
		if msg.User != "" {
			t.Error("Wrong user")
		}
		if msg.BanDuration != 0 {
			t.Error("Wrong ban duration")
		}
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("First ClearChat Message unsuccessfully parsed")
	}
	bytes = []byte("@ban-duration=10 :tmi.twitch.tv CLEARCHAT #dallas :ronni")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*ClearChat); ok {
		if msg.Channel != "dallas" {
			t.Error("Wrong Channel: " + msg.Channel)
		}
		if msg.Host != "tmi.twitch.tv" {
			t.Error("Wrong host")
		}
		if msg.User != "ronni" {
			t.Error("Wrong user")
		}
		if msg.BanDuration != 10 {
			t.Error("Wrong ban duration")
		}
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("First ClearChat Message unsuccessfully parsed")
	}

	// "CLEARMSG":        CLEARMSG,
	bytes = []byte("@login=ronni;target-msg-id=abc-123-def :tmi.twitch.tv CLEARMSG #dallas :HeyGuys it's-a Me, Mario!")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*ClearMsg); ok {
		if msg.Channel != "dallas" {
			t.Error("Wrong Channel: " + msg.Channel)
		}
		if msg.Host != "tmi.twitch.tv" {
			t.Error("Wrong host")
		}
		if msg.Login != "ronni" {
			t.Error("Wrong login")
		}
		if msg.TargetMsgId != "abc-123-def" {
			t.Error("Wrong ban duration")
		}
		if msg.Message != "HeyGuys it's-a Me, Mario!" {
			t.Error("Wrong message")
		}
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("First ClearChat Message unsuccessfully parsed")
	}

	// "GLOBALUSERSTATE": GLOBALUSERSTATE,
	bytes = []byte("@badge-info=subscriber/8;badges=subscriber/6;color=#0D4200;display-name=ronni;emote-sets=0,33,50,237,793,2126,3517,4578,5569,9400,10337,12239;turbo=0;user-id=1337;user-type=admin :tmi.twitch.tv GLOBALUSERSTATE")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*GlobalUserState); ok {
		if msg.DisplayName != "ronni" {
			t.Error("Wrong display name")
		}
		// @TODO more tags
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("Join Message unsuccessfully parsed")
	}

	// "HOSTTARGET":      HOSTTARGET,

	bytes = []byte(":tmi.twitch.tv HOSTTARGET #dallas :ronni 1600")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*HostTarget); ok {
		if msg.Channel != "dallas" {
			t.Error("Wrong Channel: " + msg.Channel)
		}
		if msg.Host != "tmi.twitch.tv" {
			t.Error("Wrong host")
		}
		if msg.TargetChannel != "ronni" {
			t.Error("Wrong target host")
		}
		if msg.NumberOfViewers != 1600 {
			t.Error("Wrong number of viewers")
		}
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("Join Message unsuccessfully parsed")
	}

	// "JOIN":            JOIN,
	// Nothing special to check with the JOIN message. Just make sure its type is right
	bytes = []byte(":ronni!ronni@ronni.tmi.twitch.tv JOIN #dallas")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*Join); ok {
		if msg.Channel != "dallas" {
			t.Error("Wrong Channel: " + msg.Channel)
		}
		if msg.Host != "ronni.tmi.twitch.tv" {
			t.Error("Wrong host")
		}
		if msg.Nickname != "ronni" {
			t.Error("Wrong nickname")
		}
		if msg.User != "ronni" {
			t.Error("Wrong username")
		}
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("Join Message unsuccessfully parsed")
	}

	// "NOTICE":          NOTICE,
	bytes = []byte("@msg-id=slow_off :tmi.twitch.tv NOTICE #dallas :This room is no longer in slow mode.")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*Notice); ok {
		if msg.Channel != "dallas" {
			t.Error("Wrong Channel: " + msg.Channel)
		}
		if msg.Message != "This room is no longer in slow mode." {
			t.Error("Wrong Message")
		}
		if msg.MsgId != "slow_off" {
			t.Error("Wrong msg id")
		}
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("Notice unsuccessfully parsed")
	}

	// "PART":            PART,
	bytes = []byte(":ronni!ronni@ronni.tmi.twitch.tv PART #dallas")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*Part); ok {
		if msg.Channel != "dallas" {
			t.Error("Wrong Channel: " + msg.Channel)
		}
		if msg.Host != "ronni.tmi.twitch.tv" {
			t.Error("Wrong host")
		}
		if msg.Nickname != "ronni" {
			t.Error("Wrong nickname")
		}
		if msg.User != "ronni" {
			t.Error("Wrong username")
		}
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("Join Message unsuccessfully parsed")
	}

	// "PING":            PING,
	bytes = []byte("PING :tmi.twitch.tv")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*Ping); ok {
		if len(msg.Servers) != 1 {
			t.Error("Wrong server number")
		}
		if msg.Servers[0] != ":tmi.twitch.tv" {
			t.Error("Wrong server")
		}
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("Ping Message unsuccessfully parsed")
	}

	// "PRIVMSG":         PRIVMSG,
	bytes = []byte("@badge-info=;badges=global_mod/1,turbo/1;color=#0D4200;display-name=ronni;emotes=25:0-4,12-16/1902:6-10;id=b34ccfc7-4977-403a-8a94-33c6bac34fb8;mod=0;room-id=1337;subscriber=0;tmi-sent-ts=1507246572675;turbo=1;user-id=1337;user-type=global_mod :ronni!ronni@ronni.tmi.twitch.tv PRIVMSG #dallas :Kappa Keepo Kappa")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*PrivMsg); ok {

		// @TODO parse/verify all tags

		if msg.Channel != "dallas" {
			t.Error("Wrong channel")
		}
		if msg.User != "ronni" {
			t.Error("Wrong user")
		}
		if msg.Message != "Kappa Keepo Kappa" {
			t.Error("Wrong message")
		}

	} else {
		fmt.Printf("%T\n", msg)
		t.Error("PrivMsg Message unsuccessfully parsed")
	}

	// "RECONNECT":       RECONNECT,

	// "ROOMSTATE":       ROOMSTATE,
	bytes = []byte("@emote-only=0;followers-only=0;r9k=0;slow=10;subs-only=0 :tmi.twitch.tv ROOMSTATE #dallas")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*RoomState); ok {
		if msg.EmoteOnly {
			t.Error("Wrong emote only bool")
		}
		if msg.R9K {
			t.Error("Wrong r9k")
		}
		if msg.SubsOnly {
			t.Error("Wrong subs only")
		}
		if msg.FollowersOnly != 0 {
			t.Error("Wrong followers only")
		}
		if msg.Slow != 10 {
			t.Error("Wrong slow value")
		}
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("Roomstate Message unsuccessfully parsed")
	}

	// "USERNOTICE":      USERNOTICE,
	bytes = []byte("@badge-info=;badges=staff/1,broadcaster/1,turbo/1;color=#008000;display-name=ronni;emotes=;id=db25007f-7a18-43eb-9379-80131e44d633;login=ronni;mod=0;msg-id=resub;msg-param-cumulative-months=6;msg-param-streak-months=2;msg-param-should-share-streak=1;msg-param-sub-plan=Prime;msg-param-sub-plan-name=Prime;room-id=1337;subscriber=1;system-msg=ronni\\shas\\ssubscribed\\sfor\\s6\\smonths!;tmi-sent-ts=1507246572675;turbo=1;user-id=1337;user-type=staff :tmi.twitch.tv USERNOTICE #dallas :Great stream -- keep it up!")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*UserNotice); ok {
		if msg.Channel != "dallas" {
			t.Error("Wrong channel")
		}
		if msg.Message != "Great stream -- keep it up!" {
			t.Error("Wrong Message")
		}
		if msg.Login != "ronni" {
			t.Error("Wrong login")
		}
		if msg.DisplayName != "ronni" {
			t.Error("Wrong Displayname")
		}
		// @TODO Test rest of tags?
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("Usernotice Message unsuccessfully parsed")
	}

	// "USERSTATE":       USERSTATE,
	bytes = []byte("@badge-info=;badges=staff/1;color=#0D4200;display-name=ronni;emote-sets=0,33,50,237,793,2126,3517,4578,5569,9400,10337,12239;mod=1;subscriber=1;turbo=1;user-type=staff :tmi.twitch.tv USERSTATE #dallas")
	ircMsg = bytesToIrcMessage(bytes)
	if msg, ok := ircMsg.(*UserState); ok {
		if msg.Channel != "dallas" {
			t.Error("Wrong channel")
		}
		if msg.DisplayName != "ronni" {
			t.Error("Wrong display name")
		}
		if !msg.Subscriber {
			t.Error("Wrong Subscriber")
		}
		// @TODO Test rest of tags?
	} else {
		fmt.Printf("%T\n", msg)
		t.Error("Usernotice Message unsuccessfully parsed")
	}
}
