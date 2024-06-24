package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/vorbis"
	"github.com/gopxl/beep/wav"
	"github.com/h2non/filetype"
)

var firstLoad = true
var fadeLoopInterval = 50
var minVolume = -6.0

/*
func stopAll() { //TODO

		// Pause and stop all playbacks
		for _, v := range playbacks {
			v.Control.Paused = true
			v.Control.Streamer = nil
			//SetPlayState(v.File, false)
		}
		speaker.Clear() // Clear the speaker and make it shut up

		// Reset the map
		playbacks = make(map[int]playback)

}
*/

func getStartFinalVals(direction int) (float64, float64) {
	// returns startVal, finalVal
	switch {
	case direction < 0:
		return 0.0, minVolume
	case direction > 0:
		return minVolume, 0.0
	}
	return 0.0, 0.0
}

func doFade(audiofile *AudioFile, direction int, duration int, stop bool) {
	startVal, finalVal := getStartFinalVals(direction)

	i := 0
	fmt.Println("FadeStart: ", time.Now())

	// Calculate the exponent for the exponential fade
	//exponent := 2.0 // You can adjust this value for the desired curve
	exponent := 0.9
	// Determine the direction of the fade
	if direction > 0 {
		for i = 0; i < duration/fadeLoopInterval; i++ {
			// Safety
			if audiofile.Volume.Volume > finalVal {
				audiofile.Volume.Volume = 0
				break
			}

			// Calculate the step using the exponential function
			step := (finalVal - startVal) * math.Pow(float64(i)/(float64(duration/fadeLoopInterval)), exponent)

			speaker.Lock()
			audiofile.Volume.Volume = startVal + step
			speaker.Unlock()
			//fmt.Println("Step ", i, ": ", step)

			time.Sleep(time.Duration(fadeLoopInterval) * time.Millisecond)
		}
		audiofile.Volume.Volume = 0
	} else if direction < 0 {
		for i = 0; i < duration/fadeLoopInterval; i++ {
			// Safety
			if audiofile.Volume.Volume < finalVal {
				audiofile.Volume.Volume = minVolume
				break
			}

			// Calculate the step using the exponential function
			step := (finalVal - startVal) * (1 - math.Pow(1-float64(i)/(float64(duration/fadeLoopInterval)), exponent))

			speaker.Lock()
			audiofile.Volume.Volume = startVal + step
			speaker.Unlock()

			time.Sleep(time.Duration(fadeLoopInterval) * time.Millisecond)
		}
		audiofile.Volume.Volume = minVolume
	}

	fmt.Println("FadeEnd: ", time.Now())

	if stop {
		StopSound(audiofile)
	}
}

func FadeoutStopAll(fadetime int) {
	for _, k := range Playing {
		if k.PlayState {
			go FadeoutStop(k, fadetime)
		}
	}
}
func FadeoutStop(audiofile *AudioFile, fadetime int) {
	doFade(audiofile, -1, fadetime, true)
}

func GetPlaybacksByCode(code string) ([]*AudioFile, error) {
	var audiofiles []*AudioFile

	for _, audiofile := range Playing {
		if audiofile.Parent.Code == code || audiofile.Code == code {
			audiofiles = append(audiofiles, audiofile)
		}
	}

	if len(audiofiles) < 1 {
		return audiofiles, errors.New("no audiofile found with this code")
	}

	return audiofiles, nil
}

func GetAudioFileByCode(code string) (*AudioFile, error) {
	var audiofile AudioFile
	idx := slices.IndexFunc(AudioFiles, func(c *AudioFile) bool { return c.Code == code })
	if idx >= 0 {
		return AudioFiles[idx], nil
	}
	return &audiofile, errors.ErrUnsupported

}

func StopSound(audiofile *AudioFile) {
	fmt.Println("Stop Sound called")
	audiofile.Control.Paused = true
	audiofile.Control.Streamer = nil
	audiofile.PlayState = false

	idx := slices.IndexFunc(Playing, func(c *AudioFile) bool { return c.Code == audiofile.Code })
	if idx >= 0 {
		playingmut.Lock()
		Playing = append(Playing[:idx], Playing[idx+1:]...)
		playingmut.Unlock()
	}

}

func DeleteAudioFile(audiofile *AudioFile) {
	//StopSound(audiofile)
	name := audiofile.Name
	idx := slices.IndexFunc(AudioFiles, func(c *AudioFile) bool { return c == audiofile })
	if idx >= 0 {
		AudioFiles = append(AudioFiles[:idx], AudioFiles[idx+1:]...)
	}
	os.Remove(path.Join(conf.SoundFolder, name))

}

func PauseSound(audiofile *AudioFile) {
	if audiofile.PlayState {
		audiofile.Control.Paused = true
	}
}
func PlaySound(audiofile *AudioFile) {
	audiofile.Control.Paused = false
}
func StartSound(sourcefile *AudioFile, loop bool, fadeInTime int, code string) {
	PlayCount += 1
	audiofile := new(AudioFile)
	*audiofile = *sourcefile
	playingmut.Lock()

	//Store reference to parent
	audiofile.Parent = sourcefile
	Playing = append(Playing, audiofile)
	playingmut.Unlock()
	audiofile.Code = uuid.New().String()
	fmt.Println("Playing sound: " + audiofile.Name)
	audiofile.Id = PlayCount
	fmt.Println("Trying to play sound")
	audiofile.PlayState = true
	amountOfLoops := 1
	if loop {
		amountOfLoops = -1
		fmt.Println("Looping sound: " + audiofile.Name)
	}
	audiofile.Seeker = audiofile.Buffer.Streamer(0, audiofile.Buffer.Len())
	volume := float64(0)
	fade := false
	if fadeInTime > 0 {
		volume = minVolume
		fade = true
	}
	done := make(chan bool)
	loopStreamer := beep.Loop(amountOfLoops, audiofile.Seeker)
	audiofile.Control = &beep.Ctrl{Streamer: loopStreamer, Paused: false}
	audiofile.Volume = &effects.Volume{
		Streamer: audiofile.Control,
		Base:     2,
		Volume:   volume,
		Silent:   false,
	}

	audiofile.Loop = loop
	audiofile.Done = done
	if fade {
		go doFade(audiofile, 1, fadeInTime, false)
	}
	speaker.PlayAndWait(audiofile.Volume)
	fmt.Println("Finished playing sound: " + audiofile.Name)
	StopSound(audiofile)
}
func loadAudioFile(fpath string) {
	var audiofile AudioFile
	t, err := os.Stat(fpath) // Check if the file exists
	if errors.Is(err, os.ErrNotExist) {
		return
	}

	if t.IsDir() { // Make sure it is not a folder we are trying to play
		return
	}

	file, err := os.Open(fpath)
	if err != nil {
		return
	}

	buf, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatal("Fatal error while opening: " + err.Error())
		return
	}

	kind, err := filetype.Match(buf)

	if err != nil {
		log.Fatal("Fatal error while detecting file type: " + err.Error())
		return
	}
	var streamer beep.StreamSeekCloser
	log.Println("File type: " + kind.MIME.Subtype)
	if kind.MIME.Subtype == "mpeg" {
		streamer, audiofile.Format, _ = mp3.Decode(file)
	} else if kind.MIME.Subtype == "x-wav" {
		streamer, audiofile.Format, _ = wav.Decode(file)
	} else if kind.MIME.Subtype == "x-flac" {
		streamer, audiofile.Format, _ = flac.Decode(file)
	} else if kind.MIME.Subtype == "ogg" {
		streamer, audiofile.Format, _ = vorbis.Decode(file)
	} else {
		fmt.Println("!!!!! Unsupported file type for " + file.Name())
		return
	}
	if firstLoad {
		speaker.Init(audiofile.Format.SampleRate, audiofile.Format.SampleRate.N(time.Second/10))
		firstLoad = false
	}

	audiofile.Buffer = beep.NewBuffer(audiofile.Format)
	audiofile.Buffer.Append(streamer)
	streamer.Close()
	audiofile.Streamer = streamer
	audiofile.Name = filepath.Base(file.Name())
	audiofile.BufferState = true
	audiofile.Code = GetMD5Hash(filepath.Base(file.Name()))
	//soundObj.Url = "/v1/play?file=" + soundObj.Code
	AudioFiles = append(AudioFiles, &audiofile)
	audiofile.Done = make(chan bool)

	var response []WebsocketLibrary
	var x WebsocketLibrary
	x.Name = audiofile.Name
	x.Code = audiofile.Code

	response = append(response, x)
	SendToWS(response, &librarymsgqueue, libmsgqueuemut)
}

func loadAudioFiles() {
	AudioFiles = make([]*AudioFile, 0)
	fmt.Println("loading audio files")
	files, err := os.ReadDir(conf.SoundFolder) // Find all files in the sounds directory
	if err != nil {
		log.Fatal(err)
	}

	// Add the file data to the temp array
	for _, file := range files {
		go loadAudioFile(path.Join(conf.SoundFolder, file.Name()))
		time.Sleep(200 * time.Millisecond)
	}
}

func SendToWS[T any](response []T, queue *[]string, mut *sync.Mutex) {
	/*if len(response) < 1 {
		return
	}
	b, err := json.Marshal(response)
	if err != nil {
		fmt.Println("json marshal error")
		return
	}
	mut.Lock()
	*queue = append(*queue, string(b))
	mut.Unlock()
	fmt.Println("s2ws - " + fmt.Sprint(len(msgqueue)))
	*/
}

func Scale(x int, in_min int, in_max int, out_min int, out_max int) int {
	return (x-in_min)*(out_max-out_min)/(in_max-in_min) + out_min
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
