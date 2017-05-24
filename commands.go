package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"fmt"
)

func processCommand(s *discordgo.Session, m *discordgo.Message) {
	if len(m.Content) < 1 {
		return
	}
	parts := strings.Split(m.Content, " ")
	if len(parts) < 1 {
		return
	}

	//TODO
	col := data.Soundcollections[parts[0]]
	if col != nil {
		var sound *Sound
		if len(parts) > 1 {
			sound = col.GetSound(parts[1])
		}
		if sound == nil {
			sound = col.GetRandomSound()
		}
		channel, _ := s.State.Channel(m.ChannelID)
		guild, err := s.State.Guild(channel.GuildID)
		if guild == nil || err != nil {
			return
		}
		vc := getCurrentVoiceChannel(s, m.Author, guild)
		if vc != nil {
			enqueueSound(s, vc.GuildID, vc.ID, sound)
		}
		return
	}

	if parts[0] == "stopwaiting" {
		if foreignCommand {
			s.ChannelMessageSend(m.ChannelID, "Ok, no waiting")
		}
		resetForeignCommand()
	}
	for _, c := range commands {
		if parts[0] == c.Command {
			c.Function(s, m, parts)
		}
	}
}

func resetForeignCommand() {
	foreignCommand = false
	foreignCommandType = ""
	foreignCommandUser = nil
}

func onForeignCommand(s *discordgo.Session, m *discordgo.Message) {
	switch foreignCommandType {
	case "image":
		if len(m.Attachments) > 0 {
			if len(foreignCommandUser) == 0 {
				foreignAvatarChange(s, m)
				return
			}
			for _, n := range foreignCommandUser {
				if n == m.Author.ID {
					foreignAvatarChange(s, m)
					return
				}
			}

		}
	}
}

func foreignAvatarChange(s *discordgo.Session, m *discordgo.Message) {
	err := changeAvatarByURL(s, m.Attachments[0].URL)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Damn, something went wrong.")
	}
	resetForeignCommand()
}

func addSoundToLibrary(file string, collection string, name string) {
	sound := NewSound(name, file)
	sound.Load()
	data.sclock.Lock()
	defer data.sclock.Unlock()
	col := data.Soundcollections[collection]
	if col == nil {
		col = &SoundCollection{Name: collection}
		data.Soundcollections[collection] = col
	} else {
		for _, s := range col.Sounds {
			if s.Name == name {
				name += "_d"
			}
		}
	}
	col.Sounds = append(col.Sounds, sound)
	data.changed = true
}

func Set(s *discordgo.Session, m *discordgo.Message, parts []string) {
	if m.Author.ID != OWNER {
		return
	}
	if len(parts) < 2 {
		return
	}
	switch parts[1] {
	case "avatar", "image":
		setAvatar(s, m, parts)
	}
}

func setAvatar(s *discordgo.Session, m *discordgo.Message, parts []string) {
	if len(m.Attachments) > 0 {
		err := changeAvatarByURL(s, m.Attachments[0].URL)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Sorry, I got rekt up")
		} else {
			s.ChannelMessageSend(m.ChannelID, "Done!")
		}
	} else {
		foreignCommand = true
		foreignCommandType = "image"
		if len(parts) > 2 {
			for _, mention := range m.Mentions {
				foreignCommandUser = append(foreignCommandUser, mention.ID)
			}
		}
		s.ChannelMessageSend(m.ChannelID, "Ok, waiting for an image.")
	}
}

func Get(s *discordgo.Session, m *discordgo.Message, parts []string) {
	if len(parts) < 2 {
		return
	}
	switch strings.ToLower(parts[1]) {
	case "id":
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%v's ID: %v", generateTag(m.Author.ID), m.Author.ID))
	case "owner", "master":
		WhoIsOwner(s, m, parts)
	}
}

func WhoIsOwner(s *discordgo.Session, m *discordgo.Message, parts []string) {
	if m.Author.ID == OWNER {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You are my master %v", generateTag(OWNER)))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("My master is %v", generateTag(OWNER)))
	}
}

func Kill(s *discordgo.Session, m *discordgo.Message, parts []string) {
	if m.Author.ID == OWNER {
		s.ChannelMessageSend(m.ChannelID, "Yes, Sir!\nI will kill myself.")
		saveSettings()
		exit <- true
	}
}

func PlayGame(s *discordgo.Session, m *discordgo.Message, parts []string) {
	if len(parts) > 1 {
		s.UpdateStatus(0, generateString(parts[1:]...))
	}
}

func SoundsHelp(s *discordgo.Session, m *discordgo.Message, parts []string) {
	output := "I can play these collections:\n"
	for _, c := range data.Soundcollections {
		output += fmt.Sprintf("-!%v\n", c.Name)
	}
	s.ChannelMessageSend(m.ChannelID, output)
}

func addCommand(cmd string, function commandFunc) {
	commands = append(commands, command{Command: cmd, Function: function})
}

func AddSound(s *discordgo.Session, m *discordgo.Message, parts []string) {
	if m.Author.ID != OWNER {
		return
	}
	if len(m.Attachments) > 0 {
		if len(parts) > 2 {
			file, err := GetSoundByURL(m.Attachments[0].URL, parts[1], parts[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Sorry, I got rekt up")
			} else {
				addSoundToLibrary(file, parts[1], parts[2])
				s.ChannelMessageSend(m.ChannelID, "Done!")
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "Wrong Parameter")
		}
	} else {
		//		foreignCommand = true
		//		foreignCommandType = "sound"
		//		if len(parts) > 2 {
		//			for _, mention := range m.Mentions {
		//				foreignCommandUser = append(foreignCommandUser, mention.ID)
		//			}
		//		}
		//		s.ChannelMessageSend(m.ChannelID, "Ok, waiting for an image.")
		s.ChannelMessageSend(m.ChannelID, "You forgot the file...")
	}
}
