package util

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tables := []struct {
		x float64
		y int
		n float64
	}{
		{1.41, 1, 1.4},
		{1.98, 1, 1.9},
		{2, 1, 2.0},
		{0.000140, 8, 0.000140},
		{5.567, 0, 5},
		{0.004, 2, 0},
	}

	for _, table := range tables {
		total := Truncate(table.x, table.y)
		if total != table.n {
			t.Errorf("Truncate(%.8f, %d) failed, got: %.8f, want: %.8f.", table.x, table.y, total, table.n)
		}
	}
}

// func testCalcSellPrice(t *testing.T) {
// 	bids := goup.DepthRecords{
// 		DepthRecord{Price: 0.000333, Amount: 1},
// 		DepthRecord{Price: 0.000332, Amount: 1},
// 		DepthRecord{Price: 0.000331, Amount: 1},
// 		DepthRecord{Price: 0.00033, Amount: 77990.4198},
// 		DepthRecord{Price: 0.000329, Amount: 17657.553},
// 		DepthRecord{Price: 0.000328, Amount: 34276.173},
// 		DepthRecord{Price: 0.000327, Amount: 12000.1816},
// 		DepthRecord{Price: 0.000326, Amount: 5951.4242},
// 		DepthRecord{Price: 0.000325, Amount: 15847.1916},
// 		DepthRecord{Price: 0.000324, Amount: 2876.9108},
// 		DepthRecord{Price: 0.000323, Amount: 10809.7363},
// 	}
// }
