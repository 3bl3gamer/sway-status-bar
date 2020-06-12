package filenum

import (
	"math"
	"sway_status_bar/utils"
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
