package diskstats

import (
	"fmt"
	"os"
	"sway_status_bar/utils"
	"time"
)

type DiskStatUnit struct {
	DevName         string
	SectorsRead     int64
	SectorsWritten  int64
	ReadSpeed       int64
	WriteSpeed      int64
	LastReadWriteAt time.Time
	lastUpdateAt    time.Time
}

const sectorSize = 512 //https://lkml.org/lkml/2015/8/17/269

func (u *DiskStatUnit) Update(iter int64) error {
	f, err := os.Open("/sys/class/block/" + u.DevName + "/stat")
	if err != nil {
		return err
	}
	defer f.Close()

	now := time.Now()
	timeDelta := int64(now.Sub(u.lastUpdateAt))
	u.lastUpdateAt = now

	var dummy string
	var sectorsRead, sectorsWritten int64
	fmt.Fscanf(f, "%s%s %d %s%s%s %d",
		&dummy, &dummy, &sectorsRead, &dummy, &dummy, &dummy, &sectorsWritten)

	sectorsReadDelta := utils.MaxInt64(0, sectorsRead-u.SectorsRead)
	sectorsWrittenDelta := utils.MaxInt64(0, sectorsWritten-u.SectorsWritten)

	if sectorsReadDelta > 0 || sectorsWrittenDelta > 0 {
		u.LastReadWriteAt = now
	}
	u.ReadSpeed = sectorsReadDelta * int64(time.Second) / timeDelta * sectorSize
	u.WriteSpeed = sectorsWrittenDelta * int64(time.Second) / timeDelta * sectorSize
	u.SectorsRead = sectorsRead
	u.SectorsWritten = sectorsWritten
	return nil
}
