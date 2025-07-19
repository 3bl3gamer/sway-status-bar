package file

import (
	"math"
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
