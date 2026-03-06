// nxtools
// Written by J.F. Gratton <jean-francois@famillegratton.net>
// Original timestamp: 2026/03/05 20:26
// Original filename: src/shared/helpers.go

package shared

import (
	"fmt"
	"math"
)

func FormatSize(sz int64) string {
	numSize := (float64)(sz) / 1000.0 / 1000.0 // this will give us the size in MB
	if (int)(math.Log10(float64(numSize))) > 2 {
		return fmt.Sprintf("%.3f GB", numSize/1000.0)
	} else {
		return fmt.Sprintf("%.3f MB", numSize)
	}
}
