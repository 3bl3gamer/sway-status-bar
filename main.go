package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sway_status_bar/cpuload"
	"sway_status_bar/diskstats"
	"sway_status_bar/file"
	"sway_status_bar/memory"
	"sway_status_bar/networks"
	"sway_status_bar/nvidia"
	"sway_status_bar/swaymsg"
	"sway_status_bar/utils"
	"sway_status_bar/volume"
	"time"
)

const interval = time.Second

var cpuTempFPath = utils.FindHwmonDirByNameSafe("/sys/devices/platform/coretemp.0/hwmon/hwmon*", "coretemp", onError) + "/temp1_input"
var ssdKingstonTempFPath = utils.FindHwmonDirBySubsysnqnSafe("/sys/devices/pci0000:00/0000:00:06.*/0000:0*:00.0/nvme/nvme*/hwmon*", "kingston", onError) + "/temp1_input"
var ssdSamsungDirPath = utils.FindHwmonDirBySubsysnqnSafe("/sys/devices/pci0000:00/0000:00:06.*/0000:0*:00.0/nvme/nvme*/hwmon*", "samsung", onError)
var ssdSamsungControllerTempFPath = utils.FindTempInputByLabelSafe(ssdSamsungDirPath, "Sensor 1", onError)
var ssdSamsungNANDTempFPath1 = utils.FindTempInputByLabelSafe(ssdSamsungDirPath, "Sensor 2", onError)

// var chipsetTempFPath = utils.FindHwmonFileByNameSafe("/sys/devices/virtual/thermal/thermal_zone0/hwmon*", "acpitz", onError) + "/temp1_input"
// var wifiTempFPath = utils.FindHwmonFileByNameSafe("/sys/devices/virtual/thermal/thermal_zone*/hwmon*", "iwlwifi_1", onError) + "/temp1_input"
var batDirPath = "/sys/devices/LNXSYSTM:00/LNXSYBUS:00/PNP0A08:00/device:08/PNP0C09:01/PNP0C0A:03/power_supply/BAT0"

var fansDirPath = utils.FindHwmonDirByNameSafe("/sys/devices/platform/asus-nb-wmi/hwmon/hwmon*", "asus", onError)

const (
	fiMicrochip            = "\uf2db"
	fiMemory               = "\uf538"
	fiHdd                  = "\uf0a0"
	fiFolder               = "\uf07b"
	fiKeyboard             = "\uf11c"
	fiGamepad              = "\uf11b"
	fiVolumeMute           = "\uf6a9"
	fiVolumeOff            = "\uf026"
	fiVolumeDown           = "\uf027"
	fiVolumeUp             = "\uf028"
	fiBluetooth            = "\uf293"
	fiBluetoothB           = "\uf294"
	fiNetworkWired         = "\uf6ff"
	fiEthernet             = "\uf796"
	fiWifi                 = "\uf1eb"
	fiMicrophone           = "\uf130"
	fiMicrophoneSlash      = "\uf131"
	fiMicrophoneAlt        = "\uf3c9"
	fiMicrophoneAltSlash   = "\uf539"
	fiHeadset              = "\uf590"
	fiPodcast              = "\uf2ce"
	fiBatteryFull          = "\uf240"
	fiBatteryThreeQuarters = "\uf241"
	fiBatteryHalf          = "\uf242"
	fiBatteryQuarter       = "\uf243"
	fiBatteryEmpty         = "\uf244"
	fiBolt                 = "\uf0e7"
	fiMobileScreenButton   = "\uf3cd"
	fiMountainSun          = "\ue52f"
	fiVolcano              = "\uf770"
	fiMountain             = "\uf6fc"
	fiFan                  = "\uf863"
)

const pangoSubSeparator = "<span color='#AAA'>/</span>"
const (
	dimCol      = "#888888"
	warnCol     = "#FC0"
	onCol       = "#77FF77"
	rxCol       = "#77FF77"
	rxFractCol  = "#559955"
	txCol       = "#FF7777"
	txFractCol  = "#995555"
	titleBgCol0 = "#005577"
	titleBgCol1 = "#007733"
)

const thinSpace = "\u2009"

// For some reason, when left and right borders are zero, the text goes beyond the bottom padding on the right (for a bit).
// Fixing this by adding a bit of transparent text on the right.
const blockExtraRightPad = thinSpace

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
	p(`{"full_text":` + fmt.Sprintf(format, a...) + `, "markup":"pango", "border_left":0, "border_right":0, "border_bottom":3, "border_top":0, "border":"` + underlineCol + `"},`)
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
		if e.Container.Name != "" && e.Container.Focused {
			title = e.Container.Name
			onUpdate()
		}
		if e.Input.XkbActiveLayoutName != "" {
			layoutIndex = e.Input.XkbActiveLayoutIndex
			onUpdate()
		}
	}

	cpuTemp := &file.FileNumUnit{FPath: cpuTempFPath, Denum: 1000}
	ssdKingstonTemp := &file.FileNumUnit{FPath: ssdKingstonTempFPath, Denum: 1000}
	ssdSamsungControllerTemp := &file.FileNumUnit{FPath: ssdSamsungControllerTempFPath, Denum: 1000}
	ssdSamsungNANDTemp := &file.FileNumUnit{FPath: ssdSamsungNANDTempFPath1, Denum: 1000}
	// chipsetTemp := &file.FileNumUnit{FPath: chipsetTempFPath, Denum: 1000}
	// wifiTemp := &file.FileNumUnit{FPath: wifiTempFPath, Denum: 1000}
	cpuLoad := &cpuload.CpuLoadUnit{}
	memory := &memory.MemoryUnit{}
	nvidiaGPU := &nvidia.NvidiaGPUUnit{}
	cpuFanSpeed := &file.FileNumUnit{FPath: fansDirPath + "/fan1_input", Denum: 1}
	gpuFanSpeed := &file.FileNumUnit{FPath: fansDirPath + "/fan2_input", Denum: 1}
	volume := &volume.VolumeUnit{}
	nets := &networks.NetworksUnit{SkipFunc: func(name string) bool {
		return name == "lo" || strings.HasPrefix(name, "docker") ||
			strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "veth")
	}}
	batCapacity := &file.FileNumUnit{FPath: batDirPath + "/capacity", Denum: 1}
	batMaxCapacity := &file.FileNumUnit{FPath: batDirPath + "/charge_control_end_threshold", Denum: 1}
	batPowerNow := &file.FileNumUnitNonZero{FPath: batDirPath + "/power_now", Denum: 1000000}
	diskStatKingston := &diskstats.DiskStatUnit{DevName: utils.FindDevNameByIDSafe("nvme-KINGSTON_SKC3000D2048G_50026B76865102B9", onError)}
	diskStatSamsung := &diskstats.DiskStatUnit{DevName: utils.FindDevNameByIDSafe("nvme-Samsung_SSD_990_EVO_Plus_4TB_S7U8NJ0Y501696X", onError)}
	swayMsg := &swaymsg.SwayMsgMonitorUnit{Events: []string{"window", "input"}, OnEvent: onEvent}
	units := []Unit{
		cpuTemp, ssdKingstonTemp, ssdSamsungControllerTemp, ssdSamsungNANDTemp, //chipsetTemp, wifiTemp,
		cpuLoad, memory, nvidiaGPU, volume, nets,
		cpuFanSpeed, gpuFanSpeed,
		batCapacity, batMaxCapacity, batPowerNow,
		diskStatKingston, diskStatSamsung, swayMsg}

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
			cpuTempCol = "color='" + warnCol + "'"
		}
		pBlock("#b25d57", `"%s %2d%s<span color='#BBB'>%2d</span>%s <span %s>%d</span>%s%s", "urgent":%t`,
			fiMicrochip,
			cpuLoad.CpuLoad, pangoSubSeparator, cpuLoad.MaxCoreLoad, dim("%"),
			cpuTempCol, cpuTemp.Num, dim("°C"),
			blockExtraRightPad,
			cpuIsUrgent)

		// Nvidia GPU
		if time.Since(nvidiaGPU.LastAwakeAt) < 10*time.Second {
			gpuRuntimeStatus := nvidiaGPU.RuntimeStatus
			gpuTempCol := ""
			if gpuRuntimeStatus == "active" {
				gpuRuntimeStatus = "A"
				gpuTempCol = "color='" + onCol + "'"
			} else if gpuRuntimeStatus == "suspended" {
				gpuRuntimeStatus = "S"
				gpuTempCol = "color='" + dimCol + "'"
			} else {
				if gpuRuntimeStatus == "resuming" {
					gpuRuntimeStatus = "^"
				} else if gpuRuntimeStatus == "suspending" {
					gpuRuntimeStatus = "v"
				}
				gpuTempCol = "color='" + warnCol + "'"
			}
			pBlock("#b25d57", `"%s <span %s>%s</span>%s"`, fiMountainSun, gpuTempCol, gpuRuntimeStatus, blockExtraRightPad)
		}

		// Fans
		println(cpuFanSpeed.Num)
		println(gpuFanSpeed.Num)
		print(fansDirPath)
		maxFanSpeed := cpuFanSpeed.Num
		if gpuFanSpeed.Num > maxFanSpeed {
			maxFanSpeed = gpuFanSpeed.Num
		}
		fanIconCol := ""
		if maxFanSpeed > 3000 {
			fanIconCol = "color='" + warnCol + "'"
		} else if maxFanSpeed > 2000 {
			fanIconCol = "color='" + onCol + "'"
		} else if maxFanSpeed == 0 {
			fanIconCol = "color='" + dimCol + "'"
		}
		pBlock("#b25d57", `"<span %s>%s</span>%s"`, fanIconCol, fiFan, blockExtraRightPad)

		// // Chipset
		// pBlock("#b25d57", `"%d%s %d%s"`, chipsetTemp.Num, dim("°C"), wifiTemp.Num, dim("°C"))

		// Storage
		ssdAvail := utils.AvailDiskSpaceSafe("/", onError)
		ssdIsUrgent := ssdAvail < availSpaceUrgentLimit || ssdKingstonTemp.Num > 75
		var ssdTemp string
		if ssdKingstonTemp.Num < 60 {
			ssdTemp = fmt.Sprintf("<span color='#BBB'>%d</span>%s", ssdKingstonTemp.Num, dim("°C"))
		} else {
			ssdTemp = fmt.Sprintf("%d%s", ssdKingstonTemp.Num, dim("°C"))
		}
		pBlock("#368653", `"%s %s %s%s%s %s%s", "urgent":%t`,
			fiHdd,
			utils.FmtSizePango(ssdAvail, "#FFF", "#FFF", dimCol, 0),
			utils.FmtSizePango(diskStatKingston.ReadSpeed, rxCol, rxFractCol, dimCol, 3),
			pangoSubSeparator,
			utils.FmtSizePango(diskStatKingston.WriteSpeed, txCol, txFractCol, dimCol, -3),
			ssdTemp,
			blockExtraRightPad,
			ssdIsUrgent)

		// Storage (secondary)
		ssdAvail = utils.AvailDiskSpaceSafe("/home/zblzgamer/secondary", onError)
		ssdMaxTemp := utils.MaxInt64(ssdSamsungControllerTemp.Num, ssdSamsungNANDTemp.Num)
		ssdIsUrgent = ssdAvail < availSpaceUrgentLimit || ssdMaxTemp > 75
		ssdSpeedStr := ""
		if time.Since(diskStatSamsung.LastReadWriteAt) < 10*time.Second {
			ssdSpeedStr = fmt.Sprintf(" %s%s%s",
				utils.FmtSizePango(diskStatSamsung.ReadSpeed, rxCol, rxFractCol, dimCol, 3),
				pangoSubSeparator,
				utils.FmtSizePango(diskStatSamsung.WriteSpeed, txCol, txFractCol, dimCol, -3))
		}
		if ssdSpeedStr == "" && ssdMaxTemp < 60 {
			ssdTemp = fmt.Sprintf("<span color='#BBB'>%d</span>%s", ssdMaxTemp, dim("°C"))
		} else {
			ssdTemp = fmt.Sprintf("%d%s<span color='#BBB'>%d</span>%s",
				ssdSamsungControllerTemp.Num, pangoSubSeparator, ssdSamsungNANDTemp.Num, dim("°C"))
		}
		pBlock("#368653", `"%s %s%s %s%s", "urgent":%t`,
			fiHdd,
			utils.FmtSizePango(ssdAvail, "#FFF", "#FFF", dimCol, 0),
			ssdSpeedStr,
			ssdTemp,
			blockExtraRightPad,
			ssdIsUrgent)

		// Memory
		memIsUrgent := memory.Percent >= 95
		pBlock("#4d79bb", `"%s %d%s<span color='#BBB'>%d</span>%s%s", "urgent":%t`,
			fiMemory, memory.Percent,
			pangoSubSeparator, memory.SwapPercent, dim("%"),
			blockExtraRightPad, memIsUrgent)

		// Volume
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
		pBlock("#a3609a", `"%s %2d%%%s"`, volIcon, volume.Volume, blockExtraRightPad)

		// Battery
		batIcon := fiBatteryFull
		if batCapacity.Num <= 15 {
			batIcon = fiBatteryEmpty
		} else if batCapacity.Num <= 40 {
			batIcon = fiBatteryQuarter
		} else if batCapacity.Num <= 65 {
			batIcon = fiBatteryHalf
		} else if batCapacity.Num <= 90 {
			batIcon = fiBatteryThreeQuarters
		}
		batPowerStr := ""
		if time.Since(batPowerNow.LastNonZeroAt) < 10*time.Second {
			batPowerStr = fmt.Sprintf(" %2d%s", batPowerNow.Num, dim("W"))
		}
		pBlock("#a3609a", `"%s %d%s<span color='#BBB'>%d</span>%s%s%s"`,
			batIcon, batCapacity.Num, pangoSubSeparator, batMaxCapacity.Num, dim("%"), batPowerStr, blockExtraRightPad)

		// Networks
		for _, net := range nets.Networks {
			if net.DevName == "lo" {
				continue
			}
			if time.Since(net.LastDataAt) > 10*time.Second {
				continue
			}

			name := net.DevName + ":"
			if net.DevName == "eno2" {
				name = fiEthernet
			} else if net.DevName == "ppp0" {
				name = fiNetworkWired
			} else if net.DevName == "wlo1" {
				name = fiWifi
			} else if strings.HasPrefix(net.DevName, "enp0s20f") {
				name = fiMobileScreenButton
			}
			pBlock("#86762f", `"%s %s%s%s%s"`,
				name,
				utils.FmtSizePango(net.RxSpeed, rxCol, rxFractCol, dimCol, 3),
				pangoSubSeparator,
				utils.FmtSizePango(net.TxSpeed, txCol, txFractCol, dimCol, -3),
				blockExtraRightPad)

		}

		// Time
		p(`{"full_text":"` + time.Now().Format("15:04:05") + blockExtraRightPad + `"},`)
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
		if lastErr == nil {
			for _, unit := range units {
				if updater, ok := unit.(Updater); ok {
					if err := updater.Update(iterCount); err != nil {
						onError(err)
					}
				}
			}
		}
		render()
		iterCount += 1
		time.Sleep(interval - time.Duration(time.Now().UnixNano()%int64(interval)))
	}
}
