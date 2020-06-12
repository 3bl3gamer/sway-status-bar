package utils

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
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

func FileReadInt64(fpath string) (int64, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
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
