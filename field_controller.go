package main

import "github.com/metal-tile/land/firedb"

// ConvertXYToRowCol XY座標からマップの座標を割り出す
// x -> col
// y -> row
func ConvertXYToRowCol(x float64, y float64, scale float64) (row int, col int) {
	col = int(x / (firedb.MapChipWidth * scale))
	row = int(y / (firedb.MapChipHeight * scale))

	return
}
