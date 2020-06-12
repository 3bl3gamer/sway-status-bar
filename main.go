package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sway_status_bar/cpuload"
	"sway_status_bar/diskstats"
	"sway_status_bar/filenum"
	"sway_status_bar/networks"
	"sway_status_bar/swaymsg"
	"sway_status_bar/utils"
	"sway_status_bar/volume"
	"time"
)

const interval = time.Second

const cpu_temp_fpath = "/sys/devices/platform/coretemp.0/hwmon/hwmon4/temp1_input"                  //name: coretemp
const ssd_temp_fpath = "/sys/devices/pci0000:00/0000:00:1d.0/0000:04:00.0/hwmon/hwmon1/temp1_input" //name: nvme
const chipset_temp_fpath = "/sys/devices/virtual/thermal/thermal_zone2/hwmon2/temp1_input"          //name: pch_cannonlake

const (
	fiMicrochip          = "\uf2db"
	fiMemory             = "\uf538"
	fiHdd                = "\uf0a0"
	fiFolder             = "\uf07b"
	fiKeyboard           = "\uf11c"
	fiGamepad            = "\uf11b"
	fiVolumeMute         = "\uf6a9"
	fiVolumeOff          = "\uf026"
	fiVolumeDown         = "\uf027"
	fiVolumeUp           = "\uf028"
	fiBluetooth          = "\uf293"
	fiBluetoothB         = "\uf294"
	fiNetworkWired       = "\uf6ff"
	fiEthernet           = "\uf796"
	fiWifi               = "\uf1eb"
	fiMicrophone         = "\uf130"
	fiMicrophoneSlash    = "\uf131"
	fiMicrophoneAlt      = "\uf3c9"
	fiMicrophoneAltSlash = "\uf539"
	fiHeadset            = "\uf590"
	fiPodcast            = "\uf2ce"
)

const pangoSubSeparator = "<span color='#AAA'>/</span>"
const (
	dimCol      = "#888888"
	rxCol       = "#77FF77"
	rxFractCol  = "#559955"
	txCol       = "#FF7777"
	txFractCol  = "#995555"
	titleBgCol0 = "#005577"
	titleBgCol1 = "#007733"
)

func dim(text string) string {
	return "<span color='" + dimCol + "'>" + text + "</span>"
}

const availSpaceUrgentLimit = int64(1024 * 1024 * 1024)

// useful:
//   https://fontawesome.com/icons?d=gallery&m=free
//   https://css.land/lch/
//   https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats
//   default dev: ip route show to 0.0.0.0/0

// TODO: stop_signal  cont_signal  click_events

func p(text string) {
	os.Stdout.Write([]byte(text))
}

func pBlock(underlineCol, format string, a ...interface{}) {
	p(`{"full_text":` + fmt.Sprintf(format, a...) + `, "markup":"pango", "border_left":0, "border_right":0, "border_bottom":2, "border_top":0, "border":"` + underlineCol + `"},`)
}

var iterCount = int64(0)
var lastErr error

func onError(err error) {
	lastErr = err
}

type Unit interface{}
type Updater interface {
	Update(iter int64) error
}
type Daemon interface {
	StartInBackground(onUpdate func(), onError func(error))
}

func main() {
	title := ""
	layoutIndex := int64(0)
	onEvent := func(e *swaymsg.Event, onUpdate func()) {
		if e.Container.Name != "" {
			title = e.Container.Name
			onUpdate()
		}
		if e.Input.XkbActiveLayoutName != "" {
			layoutIndex = e.Input.XkbActiveLayoutIndex
			onUpdate()
		}
	}

	cpuTemp := &filenum.FileNumUnit{FPath: cpu_temp_fpath, Denum: 1000}
	ssdTemp := &filenum.FileNumUnit{FPath: ssd_temp_fpath, Denum: 1000}
	chipsetTemp := &filenum.FileNumUnit{FPath: chipset_temp_fpath, Denum: 1000}
	cpuLoad := &cpuload.CpuLoadUnit{}
	volume := &volume.VolumeUnit{}
	nets := &networks.NetworksUnit{SkipFunc: func(name string) bool {
		return name == "lo" || strings.HasPrefix(name, "docker") ||
			strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "veth")
	}}
	diskStat := &diskstats.DiskStatUnit{DevName: "nvme0n1"}
	swayMsg := &swaymsg.SwayMsgMonitorUnit{Events: []string{"window", "input"}, OnEvent: onEvent}
	units := []Unit{cpuTemp, ssdTemp, chipsetTemp, cpuLoad, volume, nets, diskStat, swayMsg}

	renderBlocks := func() {
		// pBlock("#FFFFFF", `"i%d %d"`, iterCount, layoutIndex)

		// window title
		titleBuf, _ := json.Marshal(title)
		titleBgCol := titleBgCol0
		if layoutIndex == 1 {
			titleBgCol = titleBgCol1
		}
		p(`{"full_text":` + string(titleBuf) + `, "align":"center", "min_width":800, "background":"` + titleBgCol + `"},`)

		// CPU
		cpuIsUrgent := false
		cpuTempCol := ""
		if cpuTemp.Num >= 80 {
			cpuTempCol = "color='#F99'"
			cpuIsUrgent = true
		} else if cpuTemp.Num >= 70 {
			cpuTempCol = "color='#FC0'"
		}
		pBlock("#b25d57", `"%s %2d%s<span color='#BBB'>%2d</span>%s <span %s>%d</span>%s", "urgent":%t`,
			fiMicrochip,
			cpuLoad.CpuLoad, pangoSubSeparator, cpuLoad.MaxCoreLoad, dim("%"),
			cpuTempCol, cpuTemp.Num, dim("°C"),
			cpuIsUrgent)

		// Cchipset
		pBlock("#b25d57", `"%d%s"`, chipsetTemp.Num, dim("°C"))

		// storage
		availRoot := utils.AvailDiskSpaceSafe("/", onError)
		availHome := utils.AvailDiskSpaceSafe("/home", onError)
		availIsUrgent := availRoot < availSpaceUrgentLimit || availHome < availSpaceUrgentLimit
		pBlock("#368653", `"%s %s %s", "urgent":%t`,
			fiFolder,
			utils.FmtSizePango(availRoot, "#FFF", "#FFF", dimCol, 0),
			utils.FmtSizePango(availHome, "#FFF", "#FFF", dimCol, 0),
			availIsUrgent)
		pBlock("#368653", `"%s %s%s%s %d%s"`,
			fiHdd,
			utils.FmtSizePango(diskStat.ReadSpeed, rxCol, rxFractCol, dimCol, 3),
			pangoSubSeparator,
			utils.FmtSizePango(diskStat.WriteSpeed, txCol, txFractCol, dimCol, -3),
			ssdTemp.Num, dim("°C"))

		// volume
		volIcon := fiVolumeMute
		if !volume.IsMuted {
			if volume.Volume <= 5 {
				volIcon = fiVolumeOff
			} else if volume.Volume <= 15 {
				volIcon = fiVolumeDown
			} else {
				volIcon = fiVolumeUp
			}
		}
		pBlock("#4d79bb", `"%s %2d%%"`, volIcon, volume.Volume)

		// networks
		_, hasPpp := nets.NetworkByDevName["ppp0"]
		for _, net := range nets.ActiveNetworks {
			if net.DevName != "lo" && (net.DevName != "eno2" || !hasPpp) {
				name := net.DevName + ":"
				if net.DevName == "eth2" {
					name = fiEthernet
				} else if net.DevName == "ppp0" {
					name = fiNetworkWired
				} else if net.DevName == "wlo1" {
					name = fiWifi
				}
				pBlock("#86762f", `"%s %s%s%s"`,
					name,
					utils.FmtSizePango(net.RxSpeed, rxCol, rxFractCol, dimCol, 3),
					pangoSubSeparator,
					utils.FmtSizePango(net.TxSpeed, txCol, txFractCol, dimCol, -3))
			}
		}

		// time
		p(`{"full_text":"` + time.Now().Format("15:04:05") + `"},`)
	}
	render := func() {
		p("[")
		if lastErr == nil {
			renderBlocks()
		} else {
			buf, _ := json.Marshal(lastErr.Error())
			p(`{"full_text":` + string(buf) + `},`)
		}
		p("],\n")
	}

	onUpdate := func() { render() }

	for _, unit := range units {
		if daemon, ok := unit.(Daemon); ok {
			daemon.StartInBackground(onUpdate, onError)
		}
	}

	fmt.Println(`{"version":1}`)
	fmt.Print(`[`)
	for {
		for _, unit := range units {
			if updater, ok := unit.(Updater); ok {
				if err := updater.Update(iterCount); err != nil {
					onError(err)
				}
			}
		}
		render()
		iterCount += 1
		time.Sleep(interval - time.Duration(time.Now().UnixNano()%int64(interval)))
	}
}
