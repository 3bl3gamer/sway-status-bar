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
	// wpctl may return empty response just after sway start
	for i := 0; i < 5; i++ {
		err, hasUpdated := u.updateVolume()
		if err != nil {
			onError(err)
			break
		}
		if hasUpdated {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	go u.listenVolumeUpdates(onUpdate, onError)
}

func (u *VolumeUnit) listenVolumeUpdates(onUpdate func(), onError func(error)) {
	for {
		cmd, scanner, err := utils.ScanCommand("pactl", "subscribe")
		if err != nil {
			onError(err)
			time.Sleep(time.Second)
			continue
		}

		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "Event 'change' on sink") {
				err, hasUpdated := u.updateVolume()
				if err != nil {
					onError(err)
					continue
				}
				if hasUpdated {
					onUpdate()
				}
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

func (u *VolumeUnit) updateVolume() (error, bool) {
	cmd, scanner, err := utils.ScanCommand("wpctl", "get-volume", "@DEFAULT_AUDIO_SINK@")
	if err != nil {
		return err, false
	}

	hasReadLine := scanner.Scan()

	if err := cmd.Process.Kill(); err != nil {
		return err, false
	}

	if !hasReadLine {
		return nil, false
	}

	hasReadVolume := false
	line := scanner.Text()
	if lineRem, found := strings.CutPrefix(line, "Volume: "); found {
		lineRem, u.IsMuted = strings.CutSuffix(lineRem, "[MUTED]")

		volumeFract, err := strconv.ParseFloat(strings.TrimSpace(lineRem), 64)
		if err == nil {
			u.Volume = int64(volumeFract * 100)
			hasReadVolume = true
		}
	}
	if !hasReadVolume {
		return errors.New("unexpected wpctl response: " + line), false
	}

	if err := scanner.Err(); err != nil {
		return err, false
	}
	return nil, true
}
