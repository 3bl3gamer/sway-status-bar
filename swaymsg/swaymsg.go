package swaymsg

import (
	"encoding/json"
	"sway_status_bar/utils"
	"time"
)

type Event struct {
	Change    string
	Container struct{ Name string }
	Input     struct {
		XkbActiveLayoutName  string `json:"xkb_active_layout_name"`
		XkbActiveLayoutIndex int64  `json:"xkb_active_layout_index"`
	}
}

type SwayMsgMonitorUnit struct {
	Events  []string
	OnEvent func(event *Event, onUpdate func())
}

func (u *SwayMsgMonitorUnit) StartInBackground(onUpdate func(), onError func(error)) {
	go u.subscribe(onUpdate, onError)
}

func (u *SwayMsgMonitorUnit) subscribe(onUpdate func(), onError func(error)) {
	events, _ := json.Marshal(u.Events)
	for {
		_, out, err := utils.ExecCommand("swaymsg", "-t", "subscribe", "-m", string(events))
		if err != nil {
			onError(err)
			time.Sleep(time.Second)
			continue
		}
		dec := json.NewDecoder(out)
		for {
			event := &Event{}
			if err := dec.Decode(event); err != nil {
				onError(err)
				time.Sleep(time.Second)
				break
			}
			u.OnEvent(event, onUpdate)
		}
	}
}
