# Sway Status Bar

To be used as status command in Sway config (`status_command /path/to/sway_status_bar`).

Example:
![screenshot](https://user-images.githubusercontent.com/1857617/84563342-78e71880-ad63-11ea-9269-d98d3171eb67.png)

## Configuring

There are no config files, so you will have to edit the source code (`main.go` mostly).

Logic consists of two main parts: `units` array and `renderBlocks()` fucntion. Units read, process and store some system data (like temperature values) which then formatted and sent to stdout in `renderBlocks`.

### Units

Unit is any struct with `Update(iter int64) error` method which will be called every second. `iter` is incremented by 1 after each call and my be used to perform some action each N second only.

If unit needs to listen and react to some events (like volume change), it should also implement `StartInBackground(onUpdate func(), onError func(error))`.

#### filenum.FileNumUnit{FPath, Denum}

Just reads integer from file `FPath`, divides it by `Denum` and saved to `Num` field. Useful for temperatures in `/sys/devices/.../hwmon1/temp1_input` stored as `celsius * 1000`.

#### cpuload.CpuLoadUnit{}

Parses `/proc/stat`, stores whole CPU load in `CpuLoad` and max single core load in `MaxCoreLoad`.

#### memory.MemoryUnit{}

Parses `/proc/meminfo`, stores `MemTotal`, `MemFree` and `MemAvailable` in `Total`, `Free` and `Available` fields. Additionaly has field `Percent = (Total - Available) * 100 / Total`.

#### volume.VolumeUnit{}

Listens for volume changes, updates `Volume` and `IsMuted` fields. Requires `libpulse` (`pactl`) and `pulseaudio` (`pacmd`).

#### networks.NetworksUnit{SkipFunc}

Gathers network interfaces stats from `/proc/net/dev`, saves them to `Networks` array and `NetworkByDevName` map.

`SkipFunc` may be used to ignore some interfeces: `SkipFunc: func(name string) bool { return name == "lo" }`

#### diskstats.DiskStatUnit{DevName}

Parses `/sys/class/block/<DevName>/stat`, updates `SectorsRead`, `SectorsWritten`, `ReadSpeed` and `WriteSpeed` (as bytes/second) fields.

#### swaymsg.SwayMsgMonitorUnit{Events, OnEvent}

Runs `swaymsg -t subscribe -m <Events>`, calls `OnEvent` for each event. Useful for updating current window title.

### renderBlocks()

Should ouput JSON-like structure (`man swaybar-protocol`). Uses `p(string)` which is shortcut to `os.Stdout.Write` and `pBlock(underlineCol, format, args)` which outputs block with underline and enabled [pango markup](https://developer.gnome.org/pygtk/stable/pango-markup-language.html). They both **do not JSON-escape arguments** (although this is almost never needed).

Icons can be added using emoji or some special fonts like font-awesome (which is currently used and requires `otf-font-awesome` package, more icons can be found at https://fontawesome.com/icons?d=gallery&m=free).
