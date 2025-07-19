package file

import (
	"bytes"
	"math"
	"os"
	"sway_status_bar/utils"
	"time"
)

type FileNumUnit struct {
	FPath string
	Num   int64
	Denum float64
}

func (u *FileNumUnit) Update(iter int64) error {
	val, err := utils.FileReadInt64(u.FPath)
	if err != nil {
		return err
	}
	u.Num = int64(math.Round(float64(val) / u.Denum))
	return nil
}

type FileNumUnitNonZero struct {
	FPath         string
	Num           int64
	Denum         float64
	LastNonZeroAt time.Time
}

func (u *FileNumUnitNonZero) Update(iter int64) error {
	val, err := utils.FileReadInt64(u.FPath)
	if err != nil {
		return err
	}
	if val != 0 {
		u.LastNonZeroAt = time.Now()
	}
	u.Num = int64(math.Round(float64(val) / u.Denum))
	return nil
}

type FileContentUnit struct {
	FPath   string
	Trim    bool
	Content string
}

func (u *FileContentUnit) Update(iter int64) error {
	buf, err := os.ReadFile(u.FPath)
	if err != nil {
		return err
	}
	if u.Trim {
		buf = bytes.TrimSpace(buf)
	}
	u.Content = string(buf)
	return nil
}
