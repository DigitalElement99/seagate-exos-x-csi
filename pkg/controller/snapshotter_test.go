package controller

import (
	"testing"

	storageapitypes "github.com/Seagate/seagate-exos-x-api-go/v2/pkg/common"
)

// Regression: total-size-numeric is in 512-byte blocks; reporting it
// as SizeBytes made every restoreSize 512x too small (100Gi -> 200Mi).
func TestNewSnapshotFromResponseSizeBytes(t *testing.T) {
	s, err := newSnapshotFromResponse(&storageapitypes.SnapshotObject{ObjectName: "snapshot", Name: "snap", TotalSizeNumeric: 209715200})
	if err != nil {
		t.Fatal(err)
	}
	if s.SizeBytes != 107374182400 {
		t.Errorf("SizeBytes = %d, want 107374182400", s.SizeBytes)
	}
}
