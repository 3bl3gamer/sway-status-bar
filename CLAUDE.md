## Build and Run

```bash
go build          # produces ./sway_status_bar binary
./sway_status_bar # run directly; outputs swaybar JSON to stdout
```

## Architecture

A long-lived Go daemon that outputs [swaybar protocol](https://man.archlinux.org/man/swaybar-protocol.7) JSON to stdout, used as `status_command` in Sway config.

**Two-phase design:**

1. **Units** — structs that read/process system data every second. Each implements:
   - `Update(iter int64) error` — called each second; use `iter` for N-second intervals
   - Optionally `StartInBackground(onUpdate func(), onError func(error))` — for event-driven units (volume, swaymsg)

2. **`renderBlocks()`** in `main.go` — reads data stored in unit fields and formats it as swaybar JSON blocks using `pBlock(underlineCol, format, args)` and Pango markup with Font Awesome icons.

**Units:**

| Package | Struct | Data source |
|---------|--------|-------------|
| `file` | `FileNumUnit{FPath, Denum}` | Reads integer from sysfs file, divides by Denum |
| `cpuload` | `CpuLoadUnit` | `/proc/stat` → `CpuLoad`, `MaxCoreLoad` |
| `memory` | `MemoryUnit` | `/proc/meminfo` → `Total`, `Available`, `Percent` |
| `volume` | `VolumeUnit` | `wpctl`/`pactl` events → `Volume`, `IsMuted` |
| `networks` | `NetworksUnit{SkipFunc}` | `/proc/net/dev` → `Networks`, `NetworkByDevName` |
| `diskstats` | `DiskStatUnit{DevName}` | `/sys/class/block/<DevName>/stat` → read/write speeds |
| `nvidia` | `NvidiaGPUUnit` | `/sys/bus/pci/.../power/runtime_status` |
| `swaymsg` | `SwayMsgMonitorUnit{Events, OnEvent}` | `swaymsg -t subscribe -m` events |

## Configuration

**There are no config files.** All customization is in `main.go`:
- Hardware paths resolved via `utils.FindHwmonDirByName*()` and `utils.FindDevNameByIDSafe()`
- Colors and thresholds as constants at the top of `main.go`
- Block content and layout in `renderBlocks()`

This codebase is tailored to specific ASUS G15 2021 (G15 branch) and M16 2022 (M16 branch) laptops. Hardware sysfs paths, SSD names, fan directories, and device IDs are all hardcoded for these machines.

## Rendering Notes

- `p(string)` — shortcut for writing directly to stdout (no JSON escaping)
- `pBlock(underlineCol, format, args)` — outputs a swaybar block with Pango markup support
- Icons use Font Awesome (requires `otf-font-awesome`); browse at https://fontawesome.com/icons
- Pango markup docs: https://developer.gnome.org/pygtk/stable/pango-markup-language.html
