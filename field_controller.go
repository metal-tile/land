package main

const (
	// MapSizeRow is Map縦幅
	MapSizeRow = 200

	// MapSizeCol is Map横幅
	MapSizeCol = 255

	// MapChipWidth マップチップ1つの幅
	MapChipWidth = 32.0

	// MapChipHeight マップチップ1つの高さ
	MapChipHeight = 32.0
)

// Field is World Map Tile
var Field [MapSizeRow][MapSizeCol]FieldValue

// FieldValue is Fieldの1chipを表す構造体
type FieldValue struct {
	Row      int
	Col      int
	ChipID   int
	HitPoint float64
}

// ConvertXYToRowCol XY座標からマップの座標を割り出す
// x -> col
// y -> row
func ConvertXYToRowCol(x float64, y float64, scale float64) (row int, col int) {
	col = int(x / (MapChipWidth * scale))
	row = int(y / (MapChipHeight * scale))

	return
}
