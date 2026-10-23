package controller

import "testing"

// A clone is its source's size; CreateVolume grows it to the request, and a
// repeated grow or expand must be a no-op (getSizeStr(0) would ask for 4096B).
func TestBytesToGrow(t *testing.T) {
	cases := []struct{ current, required, want int64 }{
		{19998441472, 21474836480, 1476395008}, // 19072Mi source, 20Gi request
		{21474836480, 21474836480, 0},
		{21474836480, 2097152, 0}, // a restoreSize reported in blocks
		{1073741824, 0, 0},        // no capacity range
	}
	for _, c := range cases {
		if got := bytesToGrow(c.current, c.required); got != c.want {
			t.Errorf("bytesToGrow(%d, %d) = %d, want %d", c.current, c.required, got, c.want)
		}
	}
}
