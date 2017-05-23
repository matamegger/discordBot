// discordBot project main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

const (
	SETTINGS_FOLDER = "settings"
	COMMAND_FILE    = "commands.json"
	SOUNDS_FILE     = "sounds.json"
	BOT_NAME        = "JoinBot"
	MAX_QUEUE_SIZE  = 5
	BUILD           = 1
)

var (
	OWNER              string
	BASEPATH           string
	exit               chan bool
	commands           []command
	foreignCommand     bool
	foreignCommandUser []string
	foreignCommandType string
	data               Settings
	playQueue          map[string]chan *Play
	queueLock          sync.RWMutex
)

type Settings struct {
	Soundcollections map[string]*SoundCollection `json:"soundcollections"`
	Build            int                         `json:"build"`
	sclock           sync.RWMutex                `json:"-"`
	changed          bool                        `json:"-"`
}

type commandFunc func(*discordgo.Session, *discordgo.Message, []string)

type command struct {
	Command  string
	Function commandFunc
}

func (d *Settings) prepareAndLoad() {
	fmt.Println("Loading data")
	data.sclock.Lock()
	defer data.sclock.Unlock()
	if d.Soundcollections == nil {
		d.Soundcollections = make(map[string]*SoundCollection)
	} else {
		for _, c := range d.Soundcollections {
			c.Load()
		}
	}
}

func initalize() {
	ex, err := os.Executable()
	if err == nil {
		BASEPATH = filepath.Dir(ex) + string(os.PathSeparator)
	}
	addCommand("kill", Kill)
	addCommand("get", Get)
	addCommand("owner", WhoIsOwner)
	addCommand("master", WhoIsOwner)
	addCommand("set", Set)
	addCommand("addsound", AddSound)
	addCommand("sounds", SoundsHelp)
	addCommand("play", PlayGame)
	playQueue = make(map[string]chan *Play)
	loadSettings()
	data.prepareAndLoad()
}

func main() {
	var (
		Token = flag.String("t", "", "Discord Authentication Token")
		Owner = flag.String("o", "", "Owner")
		err   error
	)
	flag.Parse()

	if *Owner != "" {
		OWNER = *Owner
	} else {
		fmt.Fprintln(os.Stderr, "Owner id must be set with -o")
		return
	}

	initalize()

	discord, err := SetupDiscordConnectionAndListener(*Token)
	if err != nil {
		return
	}

	//-------------------------
	//		Closing
	//-------------------------

	// Wait for a signal to quit
	c := make(chan os.Signal, 1)
	exit = make(chan bool, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		fmt.Println("Signal: ", <-c)
		exit <- true
	}()
	<-exit

	fmt.Print("Closing connection...")
	err = discord.Close()
	if err != nil {
		fmt.Println("error:\n", err)
	}
	saveSettings()

	fmt.Println("exit")
}

//Creates a session, adds listeners and starts the session
func SetupDiscordConnectionAndListener(token string) (discord *discordgo.Session, err error) {
	discord, err = discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	// Register Handler
	discord.AddHandler(onMessageCreate)
	discord.AddHandler(onReady)
	discord.AddHandler(onGuildCreate)

	// Open the websocket and begin listening.
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
	}
	return
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	for _, s := range event.Settings.RestrictedGuilds {
		fmt.Println("rg", s)
	}

	//Set status
	if BUILD > data.Build {
		data.Build = BUILD
		data.changed = true
		s.UpdateStatus(0, "Updated to "+strconv.Itoa(BUILD))
	} else {
		s.UpdateStatus(0, "Lurking around")
	}

}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if len(m.Content) > 0 && m.Content[0] == '!' {
		m.Content = m.Content[1:]
		processCommand(s, m.Message)
	} else if len(m.Mentions) < 1 {
	}

	if foreignCommand {
		onForeignCommand(s, m.Message)
	}
}

func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	fmt.Println("guildname", event.Name)
	//TODO
}

func loadSettings() {
	cFile := BASEPATH + SETTINGS_FOLDER + string(filepath.Separator) + COMMAND_FILE
	exist, _ := exists(cFile)
	if !exist {
		return
	}
	d, err := LoadDataFromDisk(cFile)
	if err != nil {
		fmt.Println("Error Loading settings", err)
	}
	data = d
}

func saveSettings() {
	if !data.changed {
		return
	}
	cFile := BASEPATH + SETTINGS_FOLDER + string(filepath.Separator) + COMMAND_FILE
	fmt.Println("saving data at: ", cFile)
	err := SaveDataToDisk(cFile)
	if err != nil {
		fmt.Println("Error saving settings")
	}
}

func SaveDataToDisk(path string) (err error) {
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		fmt.Println("error")
		return
	}
	data.sclock.RLock()
	defer data.sclock.RUnlock()
	err = json.NewEncoder(file).Encode(data)
	return
}

func LoadDataFromDisk(path string) (d Settings, err error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		fmt.Println("error opening")
		return
	}
	err = json.NewDecoder(file).Decode(&d)
	return
}
