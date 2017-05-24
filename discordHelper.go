package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

func getCurrentVoiceChannel(s *discordgo.Session, user *discordgo.User, guild *discordgo.Guild) *discordgo.Channel {
	for _, vs := range guild.VoiceStates {
		if vs.UserID == user.ID {
			channel, _ := s.State.Channel(vs.ChannelID)
			return channel
		}
	}
	return nil
}

func changeAvatar(s *discordgo.Session, localFile string) (err error) {
	img, err := ioutil.ReadFile(localFile)
	if err != nil {
		log.Errorf("Error reading file > %s",err)
		return err
	}

	base64 := base64.StdEncoding.EncodeToString(img)

	avatar := fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(img), base64)

	fmt.Println(s.State.User.Avatar)
	u, err := s.UserUpdate("", "", BOT_NAME, avatar, "")
	if err != nil {
		log.Errorf("Error updating user avatar > %s",err)
	} else {
		if u != nil {
			s.State.User = u
		}
	}
	return
}

func changeAvatarByURL(s *discordgo.Session, url string) (err error) {

	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		log.Errorf("Error retrieving the http file > %s",err)
		return
	}

	img, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading the http response > %s",err)
		return
	}

	base64 := base64.StdEncoding.EncodeToString(img)

	avatar := fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(img), base64)

	s.State.User, err = s.UserUpdate("", "", BOT_NAME, avatar, "")
	if err != nil {
		log.Errorf("Error updating user avatar > %s",err)
	}
	return
}

//generates a mentiontag
func generateTag(userId string) string {
	return fmt.Sprintf("<@!%v>", userId)
}
