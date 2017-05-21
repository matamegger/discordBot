package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Holding multiple sounds
type SoundCollection struct {
	Name   string   `json:"name"`
	Sounds []*Sound `json:"sounds,omitempty"`
}

// Holding the name, path and data of a sound
type Sound struct {
	Name   string   `json:"name"`
	File   string   `json:"path"`
	buffer [][]byte `json:"-"`
}

// Holding information to where play which sound
type Play struct {
	sound     *Sound
	channelID string
	guildID   string
}

/**
 * Play
 */

// Creates a new Play
func NewPlay(sound *Sound, guildID, channelID string) *Play {
	return &Play{sound: sound, guildID: guildID, channelID: channelID}
}

/**
 * SoundCollection
 */

// Loads the sound data of all Sounds in the SoundCollection
func (sc *SoundCollection) Load() {
	fmt.Println("Loading Collection", sc.Name)
	for _, s := range sc.Sounds {
		s.Load()
	}
}

// Returns the Sound with the given name or nil, if no sound with
// the given name exists in this SoundCollection
func (sc *SoundCollection) GetSound(name string) *Sound {
	for _, s := range sc.Sounds {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// Returns a random sound
func (sc *SoundCollection) GetRandomSound() *Sound {
	index := rand.Intn(len(sc.Sounds))
	return sc.Sounds[index]
}

/**
 * Sound
 */

// Creats a new sound
func NewSound(name string, file string) *Sound {
	return &Sound{Name: name, File: file}
}

// Loads the Sounddata of the sound into the memory
func (s *Sound) Load() (err error) {
	fmt.Println("Loading Sound", s.Name)
	file, err := os.Open(s.File)

	if err != nil {
		fmt.Println("Error opening dca file: ", err)
		return
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			file.Close()
			if err != nil {
				return
			}
			return
		}

		if err != nil {
			fmt.Println("Error reading from dca file:", err)
			return
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file:", err)
			return
		}

		// Append encoded pcm data to the buffer.
		s.buffer = append(s.buffer, InBuf)
	}
}

/**
 * Discordbot methodes
 */

// Creates a Play and puts it into the queue of the given guild if the maximum number of
// enqued Plays is not exeeded
func enqueueSound(s *discordgo.Session, guildID, channelID string, sound *Sound) {
	play := NewPlay(sound, guildID, channelID)

	queueLock.Lock()
	defer queueLock.Unlock()
	playChannel := playQueue[guildID]

	if playChannel != nil {
		if len(playChannel) < MAX_QUEUE_SIZE {
			playChannel <- play
		}
	} else {
		playQueue[guildID] = make(chan *Play, MAX_QUEUE_SIZE)
		go playSound(s, play, nil)
	}
}

// Plays a Play in its VoiceChannel using an existing VoiceConnection or, if vc is nil
// creates its own VoiceConnection. Furthermore, if there are Plays for the same VoiceChannel,
// a new go routine will be started to play the next song
// If the VoiceConnection is connected to a different channel, the channel will be changed.
func playSound(s *discordgo.Session, play *Play, vc *discordgo.VoiceConnection) (err error) {

	if vc == nil {
		// Join the provided voice channel.
		vc, err = s.ChannelVoiceJoin(play.guildID, play.channelID, false, true)
		if err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}

	if vc.ChannelID != play.channelID {
		vc.ChangeChannel(play.channelID, false, true)
		time.Sleep(200 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)

	_ = vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range play.sound.buffer {
		vc.OpusSend <- buff
	}

	_ = vc.Speaking(false)

	queueLock.Lock()

	if len(playQueue[play.guildID]) > 0 {
		play = <-playQueue[play.guildID]
		queueLock.Unlock()
		go playSound(s, play, vc)
		return nil
	}
	delete(playQueue, play.guildID)
	queueLock.Unlock()

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	_ = vc.Disconnect()

	return nil
}
