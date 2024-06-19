package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hypebeast/go-osc/osc"
)

func oscStartServer() {
	addr := "127.0.0.1:" + fmt.Sprint(conf.OSCPort)
	d := osc.NewStandardDispatcher()
	d.AddMsgHandler("/control/start", oscStart)
	d.AddMsgHandler("/control/pause", oscPause)
	d.AddMsgHandler("/control/stop", oscStop)
	d.AddMsgHandler("/control/fadeoutstopall", oscFadeOutStopAll)

	server := &osc.Server{
		Addr:       addr,
		Dispatcher: d,
	}
	go server.ListenAndServe()
}
func oscStart(msg *osc.Message) {
	osc.PrintMessage(msg)
	if msg.CountArguments() < 1 {
		return
	}
	tt, _ := msg.TypeTags()
	var code string
	var fadetime int32 = 0
	switch tt[1:] {
	case "s": // Only File to Play, no fadetime
		code = msg.Arguments[0].(string)
	case "si": //File and fadetime
		code = msg.Arguments[0].(string)
		fadetime = msg.Arguments[1].(int32)
	case "is": //File and fadetime
		code = msg.Arguments[1].(string)
		fadetime = msg.Arguments[0].(int32)
	default:
		return
	}
	audiofile, err := GetAudioFileByCode(code)
	if err != nil {
		return
	}
	playcode := uuid.New().String()
	go StartSound(audiofile, false, int(fadetime), playcode)

	fmt.Println(code)
}

func oscPause(msg *osc.Message) {
	osc.PrintMessage(msg)
	if msg.CountArguments() < 1 {
		return
	}
	tt, _ := msg.TypeTags()
	switch tt[1:] {
	case "s":
		audiofile, err := GetAudioFileByCode(msg.Arguments[0].(string))
		if err != nil {
			return
		}
		go PauseSound(audiofile)
	default:
		return
	}

}

func oscStop(msg *osc.Message) {
	osc.PrintMessage(msg)
	if msg.CountArguments() < 1 {
		return
	}
	tt, _ := msg.TypeTags()
	switch tt[1:] {
	case "s":
		audiofile, err := GetAudioFileByCode(msg.Arguments[0].(string))
		if err != nil {
			return
		}
		go StopSound(audiofile)
	default:
		return
	}
}

func oscFadeOutStopAll(msg *osc.Message) {
	osc.PrintMessage(msg)
	tt, _ := msg.TypeTags()
	switch tt[1:] {
	case "i":
		go FadeoutStopAll(int(msg.Arguments[0].(int32)))
	default:
		go FadeoutStopAll(conf.DefaultFadeTime)
	}
}
