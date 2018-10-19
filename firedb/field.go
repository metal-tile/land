package firedb

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

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

// FieldValue is Fieldの1chipを表す構造体
type FieldValue struct {
	Row      int
	Col      int
	ChipID   int     `firestore:"chip"`
	HitPoint float64 `firestore:"hitPoint"`
}

// FieldStore is FieldStore
type FieldStore interface {
	SetValue(row int, col int, v *FieldValue) error
	GetValue(row int, col int) (*FieldValue, error)
	Watch(ctx context.Context, path string) error
}

type defaultFieldStore struct {
	mu    *sync.RWMutex
	Field [MapSizeRow][MapSizeCol]*FieldValue
}

var fieldStore FieldStore

// NewFieldStore is New FieldStore
func NewFieldStore() FieldStore {
	if fieldStore == nil {
		fieldStore = &defaultFieldStore{
			mu: &sync.RWMutex{},
		}
	}
	return fieldStore
}

// SetFieldStore is UnitTest時に実装を差し替えたいときに利用する
func SetFieldStore(s FieldStore) {
	fieldStore = s
}

func (s *defaultFieldStore) SetValue(row int, col int, v *FieldValue) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if row > len(s.Field) {
		return fmt.Errorf("row : %d > filed.length : %d", row, len(s.Field))
	}
	if col > len(s.Field[row]) {
		return fmt.Errorf("col : %d > filed[%d].length : %d", col, row, len(s.Field[row]))
	}

	s.Field[row][col] = v
	return nil
}

func (s *defaultFieldStore) GetValue(row int, col int) (*FieldValue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if row > len(s.Field) {
		return nil, fmt.Errorf("row : %d > filed.length : %d", row, len(s.Field))
	}
	if col > len(s.Field[row]) {
		return nil, fmt.Errorf("col : %d > filed[%d].length : %d", col, row, len(s.Field[row]))
	}

	return s.Field[row][col], nil
}

func (s *defaultFieldStore) Watch(ctx context.Context, path string) error {
	iter := db.Collection(path).Snapshots(ctx)
	defer iter.Stop()
	for {
		dociter, err := iter.Next()
		if err != nil {
			return err
		}
		dslist := dociter.Changes
		if err != nil {
			return err
		}
		for _, v := range dslist {
			row, col, err := buildFieldRowCol(v.Doc.Ref.ID)
			if err != nil {
				return err
			}
			var fv FieldValue
			if err := v.Doc.DataTo(&fv); err != nil {
				return err
			}
			fv.Row = row
			fv.Col = col
			if err := s.SetValue(row, col, &fv); err != nil {
				return err
			}
		}
	}
}

// buildFieldRowCol is Firestoreから送られてくるidから、FieldのRowColを抜き出す
// FieldのKeyとして `row-000-col-000` という文字列を使っているので、それをばらしている
func buildFieldRowCol(id string) (row int, col int, error error) {
	rc := strings.Split(id, "-")
	if len(rc) != 4 {
		return 0, 0, fmt.Errorf("Path has an unexpected format. id = %s", id)
	}
	const rowIndex = 1
	row, err := strconv.Atoi(rc[rowIndex])
	if err != nil {
		return 0, 0, fmt.Errorf("miss Atoi v = %s, id = %s", rc[rowIndex], id)
	}
	const colIndex = 3
	col, err = strconv.Atoi(rc[colIndex])
	if err != nil {
		return 0, 0, fmt.Errorf("miss Atoi v = %s, id = %s", rc[colIndex], id)
	}
	return row, col, nil
}
