package nvidia

import (
	"os"
	"strings"
	"time"
)

type NvidiaGPUUnit struct {
	// PowerState string
	RuntimeStatus string
	LastAwakeAt   time.Time
}

func (u *NvidiaGPUUnit) Update(iter int64) error {
	// /sys/bus/pci/devices/0000:01:00.0/power_state D0, D3Cold, ...
	// /sys/bus/pci/devices/0000:01:00.0/power/runtime_status suspended, resuming, active, ...

	// buf, err := os.ReadFile("/sys/bus/pci/devices/0000:01:00.0/power_state")
	// if err != nil {
	// 	return err
	// }
	// u.PowerState = strings.TrimSpace(string(buf))

	buf, err := os.ReadFile("/sys/bus/pci/devices/0000:01:00.0/power/runtime_status")
	if err != nil {
		return err
	}
	u.RuntimeStatus = strings.TrimSpace(string(buf))

	if u.RuntimeStatus != "suspended" {
		u.LastAwakeAt = time.Now()
	}
	return nil
}
