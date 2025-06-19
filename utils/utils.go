package utils

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func ExecCommand(name string, arg ...string) (*exec.Cmd, io.ReadCloser, error) {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM, //terminate child with parent
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	return cmd, out, nil
}

func ScanCommand(name string, arg ...string) (*exec.Cmd, *bufio.Scanner, error) {
	cmd, out, err := ExecCommand(name, arg...)
	if err != nil {
		return nil, nil, err
	}
	return cmd, bufio.NewScanner(out), nil
}

func FindHwmonDirByName(pattern, name string) (string, error) {
	fnames, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	foundFName := ""
	for _, fname := range fnames {
		buf, err := os.ReadFile(fname + "/name")
		if err != nil {
			return "", err
		}
		if strings.Trim(string(buf), "\n") == name {
			if foundFName == "" {
				foundFName = fname
			} else {
				return "", errors.New("found multiple files for " + pattern + " and " + name)
			}
		}
	}
	if foundFName == "" {
		return "", errors.New("found no files for " + pattern + " and " + name)
	}
	return foundFName, nil
}

func FindHwmonDirByNameSafe(pattern, name string, onError func(error)) string {
	fpath, err := FindHwmonDirByName(pattern, name)
	if err != nil {
		onError(err)
	}
	return fpath
}

func FindHwmonDirBySubsysnqn(pattern, subsysnqnPart string) (string, error) {
	fnames, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	foundFName := ""
	for _, fname := range fnames {
		buf, err := os.ReadFile(fname + "/device/subsysnqn")
		if err != nil {
			return "", err
		}
		if strings.Contains(string(buf), subsysnqnPart) {
			if foundFName == "" {
				foundFName = fname
			} else {
				return "", errors.New("found multiple files for " + pattern + " and subsysnq with " + subsysnqnPart)
			}
		}
	}
	if foundFName == "" {
		return "", errors.New("found no files for " + pattern + " and subsysnq with " + subsysnqnPart)
	}
	return foundFName, nil
}

func FindHwmonDirBySubsysnqnSafe(pattern, subsysnqnPart string, onError func(error)) string {
	fpath, err := FindHwmonDirBySubsysnqn(pattern, subsysnqnPart)
	if err != nil {
		onError(err)
	}
	return fpath
}

func FindTempInputByLabel(dirPath, label string) (string, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	foundFName := ""
	names, err := f.Readdirnames(-1)
	if err != nil {
		return "", err
	}
	for _, name := range names {
		if strings.HasPrefix(name, "temp") && strings.HasSuffix(name, "_label") {
			content, err := os.ReadFile(dirPath + "/" + name)
			if err != nil {
				return "", err
			}
			if strings.TrimSpace(string(content)) == label {
				if foundFName == "" {
					foundFName = name
				} else {
					return "", errors.New("found multiple temp*_label with " + label + " in " + dirPath)
				}
			}
		}
	}
	if foundFName == "" {
		return "", errors.New("found no temp*_label with " + label + " in " + dirPath)
	}
	return dirPath + "/" + strings.TrimSuffix(foundFName, "_label") + "_input", nil
}

func FindTempInputByLabelSafe(dirPath, label string, onError func(error)) string {
	fpath, err := FindTempInputByLabel(dirPath, label)
	if err != nil {
		onError(err)
	}
	return fpath
}

func FileReadInt64(fpath string) (int64, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		return 0, err
	}
	if len(buf) > 0 && buf[len(buf)-1] == '\n' {
		buf = buf[:len(buf)-1]
	}
	val, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func AvailDiskSpace(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	return int64(stat.Bavail * uint64(stat.Bsize)), nil
}

func AvailDiskSpaceSafe(path string, onError func(error)) int64 {
	avail, err := AvailDiskSpace(path)
	if err != nil {
		onError(err)
	}
	return avail
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func AbsInt64(a int64) int64 {
	if a > 0 {
		return a
	}
	return -a
}

var sizeSuffixes = []string{"B", "K", "M", "G", "T", "P"}

func FmtSizeParts(n int64) (string, string, string) {
	for i := 0; i < len(sizeSuffixes); i++ {
		if n < 1<<(10*(i+1)) || i == len(sizeSuffixes)-1 {
			// nr := n / (1 << (10 * i))
			nf := float64(n) / float64(int64(1)<<(10*i))
			if nf >= 1000 && i < len(sizeSuffixes)-1 {
				return "1.0", "", sizeSuffixes[i+1]
			} else if nf < 10 && nf > 0 {
				s := strconv.FormatFloat(nf, 'f', 1, 64)
				return s[:1], s[1:], sizeSuffixes[i]
			} else {
				return strconv.FormatInt(int64(nf), 10), "", sizeSuffixes[i]
			}
		}
	}
	return "?", "?", "?"
}

func FmtSizePango(n int64, baseCol, fracCol, letterCol string, numPadChars int64) string {
	rnd, frac, lett := FmtSizeParts(n)
	paddig := strings.Repeat(" ", int(MaxInt64(0, AbsInt64(numPadChars)-int64(len(rnd+frac)))))
	res := "<span color='" + baseCol + "'>" + rnd + "</span>"
	if frac != "" {
		res += "<span color='" + fracCol + "'>" + frac + "</span>"
	}
	res = res + "<span color='" + letterCol + "'>" + lett + "</span>"
	if numPadChars > 0 {
		res = paddig + res
	} else {
		res = res + paddig
	}
	return res
}

// nvme-KINGSTON_NUMBER_SERIAL -> nvme0n1
//
// IDs from /dev/disk/by-id/
func FindDevNameByIDSafe(id string, onError func(error)) string {
	dest, err := os.Readlink("/dev/disk/by-id/" + id)
	if err != nil {
		onError(err)
	}
	return path.Base(dest)
}
