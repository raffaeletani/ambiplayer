package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
)

var (
	AppName    = "Ambiplayer"
	AppVersion = "dev"
	AppAuthor  = "Raffaele Tani"
)

type Configuration struct {
	HttpPort        int    `json:"httpport"`
	OSCPort         int    `json:"oscport"`
	SoundFolder     string `json:"soundfolder"`
	DefaultFadeTime int    `json:"defaultfadetime"`
}

var conf = Configuration{}

type Seek struct {
	PosProgres       int
	LenProgres       int
	PosDuration      time.Duration
	LenDuration      time.Duration
	PosDisplay       string
	LenDisplay       string
	RemainingDisplay string
}
type AudioFile struct {
	Id          int
	Name        string
	Code        string
	BufferState bool
	PlayState   bool
	Seeker      beep.StreamSeeker
	Control     *beep.Ctrl
	Volume      *effects.Volume
	Loop        bool
	Format      beep.Format
	Done        chan bool
	Buffer      *beep.Buffer
	Streamer    beep.Streamer
	Seek        Seek
}

var AudioFiles []*AudioFile
var Playing []*AudioFile
var PlayCount int
var configFile string

var librarymsgqueue []string

var libmsgqueuemut *sync.Mutex
var playingmut *sync.Mutex

func setup() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dir = filepath.ToSlash(dir)
	configFile = string(path.Join(dir, "conf.json"))
	playingmut = new(sync.Mutex)
	libmsgqueuemut = new(sync.Mutex)
	fmt.Println("Welcome to " + AppName + " " + AppVersion)

	loadConfig()

	webStartServer()
	loadAudioFiles()

	oscStartServer()
}
func main() {
	systray.Run(onReady, onExit)
}
func onReady() {
	AppIcon, err := os.ReadFile("resources/icon/appicon.ico")
	println(AppIcon)
	if err != nil {
		panic(err)
	}

	systray.SetTitle(AppName)
	//systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTemplateIcon(AppIcon, AppIcon)
	systray.SetIcon(AppIcon)
	systray.SetTooltip(AppName + " " + AppVersion)
	mStopAll := systray.AddMenuItem("Stop all", "Stop all playing audio files")
	mReload := systray.AddMenuItem("Reload", "Reload all audio files")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	setup()
	/*for {
		time.Sleep(1 * time.Second)
	}*/
	for {
		select {
		case <-mStopAll.ClickedCh:
			go FadeoutStopAll(0)
		case <-mReload.ClickedCh:
			go loadAudioFiles()
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}

}

func onExit() {
	// clean up here
	fmt.Println("exit called")
	os.Exit(0)
}
