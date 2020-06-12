package cpuload

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CpuLoadUnit struct {
	CpuLoad             int64
	MaxCoreLoad         int64
	prevTotal, prevIdle []int64
}

func (u *CpuLoadUnit) Update(iter int64) error {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return err
	}
	defer f.Close()

	u.MaxCoreLoad = 0
	for {
		var name string
		fmt.Fscan(f, &name)
		isCpu := name == "cpu"
		isCore := !isCpu && strings.HasPrefix(name, "cpu")
		if !isCpu && !isCore {
			break
		}

		coreNum := int64(-1)
		if isCore {
			coreNum, err = strconv.ParseInt(name[3:], 10, 64)
			if err != nil {
				return err
			}
		}
		cpuDataIndex := coreNum + 1
		if int64(len(u.prevIdle)) <= cpuDataIndex {
			u.prevIdle = append(u.prevIdle, 0)
			u.prevTotal = append(u.prevTotal, 0)
		}

		var user, nice, system, idle, iowait, irq, softirq, steal, guest, guestNice int64
		fmt.Fscanf(f, "%d %d %d %d %d %d %d %d %d %d\n",
			&user, &nice, &system, &idle, &iowait, &irq, &softirq, &steal, &guest, &guestNice)
		total := user + nice + system + idle + iowait + irq + softirq + steal + guest + guestNice

		percent := int64(0)
		if total != u.prevTotal[cpuDataIndex] {
			percent = 100 - (idle-u.prevIdle[cpuDataIndex])*100/(total-u.prevTotal[cpuDataIndex])
		}

		u.prevTotal[cpuDataIndex] = total
		u.prevIdle[cpuDataIndex] = idle
		// fmt.Println(name, total)

		if isCpu {
			u.CpuLoad = percent
		} else {
			if percent > u.MaxCoreLoad {
				u.MaxCoreLoad = percent
			}
		}
	}
	return nil
}
