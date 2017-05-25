// discordBot project main.go
package main

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/matamegger/discordBot/logging"
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
	log                *logging.Logger
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
	log.Info("Loading ")
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
	log = logging.NewLogger("discord_bot", os.Stdout, os.Stderr)
	log.Info("Starting")
	var (
		Token = flag.String("t", "", "Discord Authentication Token")
		Owner = flag.String("o", "", "Owner")
		err   error
	)
	flag.Parse()

	if *Owner != "" {
		OWNER = *Owner
	} else {
		log.Warning("Owner ID is not set!")
		log.Notice("Set the owner ID with the -o parameter")
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
		log.Debugf("Signal: %s", <-c)
		exit <- true
	}()
	<-exit

	log.Debug("Closing discord connection...")
	err = discord.Close()
	if err != nil {
		log.Errorf("Error closing discord connection > %s", err)
	}
	saveSettings()

	log.Info("Shutdown down")
}

//Creates a session, adds listeners and starts the session
func SetupDiscordConnectionAndListener(token string) (discord *discordgo.Session, err error) {
	log.Debugf("Creating Bot with token=%s", token)
	discord, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Errorf("Error creating discord session > %s", err)
		return
	}
	// Register Handler
	discord.AddHandler(onMessageCreate)
	discord.AddHandler(onReady)
	discord.AddHandler(onGuildCreate)

	// Open the websocket and begin listening.
	err = discord.Open()
	if err != nil {
		log.Errorf("Error opening discord connection > %s", err)
	}
	return
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	//Set status
	var status string
	if BUILD > data.Build {
		data.Build = BUILD
		data.changed = true
		status = "Updated to " + strconv.Itoa(BUILD)
	} else {
		status = "Lurking around"
	}
	log.Infof("Set status to: %s", status)
	s.UpdateStatus(0, status)
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
	log.Debugf("Guild created: %s", event.Name)
	//TODO
}

func loadSettings() {
	cFile := BASEPATH + SETTINGS_FOLDER + string(filepath.Separator) + COMMAND_FILE
	exist, _ := exists(cFile)
	if !exist {
		log.Debug("Can't load settings, because the file does not exist.")
		return
	}
	var d Settings
	err := LoadObjectFromJsonFile(cFile,&d);
	if err != nil {
		log.Errorf("Error loading settings > %s", err)
	}
	data = d
}

func saveSettings() {
	if !data.changed {
		return
	}
	cFile := BASEPATH + SETTINGS_FOLDER + string(filepath.Separator) + COMMAND_FILE
	log.Debugf("Saving settings at: %s", cFile)
	data.sclock.RLock()
	defer data.sclock.RUnlock()
	err := SaveObjectAsJsonToFile(cFile, &data)
	if err != nil {
		log.Error("Error saving settings")
	}
}