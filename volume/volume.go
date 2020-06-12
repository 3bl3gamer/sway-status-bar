package volume

import (
	"errors"
	"strconv"
	"strings"
	"sway_status_bar/utils"
	"time"
)

type VolumeUnit struct {
	Volume  int64
	IsMuted bool
}

func (u *VolumeUnit) StartInBackground(onUpdate func(), onError func(error)) {
	if err := u.updateVolume(); err != nil {
		onError(err)
	}
	go u.listenVolumeUpdates(onUpdate, onError)
}

func (u *VolumeUnit) listenVolumeUpdates(onUpdate func(), onError func(error)) {
	for {
		cmd, scanner, err := utils.ScanCommand("pactl", "subscribe")
		if err != nil {
			println(1)
			onError(err)
			time.Sleep(time.Second)
			continue
		}

		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "Event 'change' on sink") {
				if err := u.updateVolume(); err != nil {
					onError(err)
					continue
				}
				onUpdate()
			}
		}
		if err := scanner.Err(); err != nil {
			onError(err)
			time.Sleep(time.Second)
			continue
		}

		// we should not be there: stdout should not end
		if err := cmd.Process.Kill(); err != nil {
			onError(err)
		}
		time.Sleep(time.Second)
	}
}

func (u *VolumeUnit) updateVolume() error {
	cmd, scanner, err := utils.ScanCommand("pacmd", "info")
	if err != nil {
		return err
	}
	prefix_defaultSnikName := "Default sink name: "
	defaultSnikName := ""
	readingDefaultSink := false
	volumeFound := false
	mutedFound := false
	for scanner.Scan() {
		line := scanner.Text()
		if readingDefaultSink && strings.HasPrefix(line, "\tvolume: ") {
			index := strings.IndexByte(line, '%')
			if index == -1 {
				return errors.New(`no '%' in line '` + line + `'`)
			}
			val, err := strconv.ParseInt(strings.TrimSpace(line[index-3:index]), 10, 64)
			if err != nil {
				return err
			}
			u.Volume = val
			volumeFound = true
		} else if readingDefaultSink && line == "\tmuted: no" {
			u.IsMuted = false
			mutedFound = true
		} else if readingDefaultSink && line == "\tmuted: yes" {
			u.IsMuted = true
			mutedFound = true
		} else if defaultSnikName == "" && strings.HasPrefix(line, prefix_defaultSnikName) {
			defaultSnikName = line[len(prefix_defaultSnikName):]
		} else if strings.HasSuffix(line, "name: <"+defaultSnikName+">") {
			readingDefaultSink = true
		}
		if volumeFound && mutedFound {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if err := cmd.Process.Kill(); err != nil {
		return err
	}
	return nil
}
