package memory

import (
	"fmt"
	"os"
)

type MemoryUnit struct {
	Total, Free, Available int64
	Percent                int64
}

func (u *MemoryUnit) Update(iter int64) error {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return err
	}
	defer f.Close()

	u.Total = -1
	u.Free = -1
	u.Available = -1
	for {
		var name string
		var size int64
		_, err = fmt.Fscanf(f, "%s %d kB\n", &name, &size)
		if err != nil {
			return err
		}
		switch name {
		case "MemTotal:":
			u.Total = size * 1024
		case "MemFree:":
			u.Free = size * 1024
		case "MemAvailable:":
			u.Available = size * 1024
		}
		if u.Total != -1 && u.Free != -1 && u.Available != -1 {
			break
		}
	}
	u.Percent = (u.Total - u.Available) * 100 / u.Total
	return nil
}
