package game

import "testing"

func TestCheckValidIndex(t *testing.T) {
	tests := []struct {
		index int
		valid bool
	}{
		{-1, false},
		{0, true},
		{1, true},
		{2, true},
		{3, false},
	}

	for _, testCase := range tests {
		if got := CheckValidIndex(testCase.index); got != testCase.valid {
			t.Errorf("CheckIndexValid(%d) = %v, want %v", testCase.index, got, testCase.valid)
		}
	}
}

func TestBoard_GetPoint(t *testing.T) {
	board := NewBoard()

	tests := []struct {
		row, col int
		want     string
		wantErr  bool
	}{
		{0, 0, "-", false},
		{1, 1, "-", false},
		{2, 2, "-", false},
		{3, 0, "", true},
		{0, 3, "", true},
	}

	for _, tt := range tests {
		got, err := board.GetPoint(tt.row, tt.col)
		if (err != nil) && !tt.wantErr {
			t.Errorf("GetPoint(%d, %d) error = %v, wantErr %v", tt.row, tt.col, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("GetPoint(%d, %d) = %v, want %v", tt.row, tt.col, got, tt.want)
		}
	}
}

func TestBoard_SetCell(t *testing.T) {
	board := NewBoard()

	tests := []struct {
		row, col int
		wantErr  bool
	}{
		{0, 0, false},
		{1, 1, false},
		{2, 2, false},
		{3, 3, true},
		{0, 0, true},
	}

	for _, tt := range tests {
		err := board.SetCell(tt.row, tt.col, "-")
		if err != nil && !tt.wantErr {
			t.Errorf("SetCell(%d, %d) error = %v, wantErr %v", tt.row, tt.col, err, tt.wantErr)
		}
	}
}

func TestBoard_CheckValidInsertion(t *testing.T) {
	board := NewBoard()

	tests := []struct {
		row, col int
		valid    bool
	}{
		{-3, 0, false},
		{0, 0, true},
		{1, 1, true},
		{2, 2, true},
		{3, 3, false},
	}

	for _, tt := range tests {
		valid := board.CheckValidInsertion(tt.row, tt.col)
		if valid != tt.valid {
			t.Errorf("CheckValidInsertion(%d, %d) = %v, want %v", tt.row, tt.col, valid, tt.valid)
		}
	}

}
