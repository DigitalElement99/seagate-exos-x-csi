package controller

import (
	"errors"
	"testing"

	storageapitypes "github.com/Seagate/seagate-exos-x-api-go/v2/pkg/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// A refused snapshot delete is HTTP 200 + Error status (err == nil) and must
// not be reported as success (#214).
func TestDeleteSnapshotError(t *testing.T) {
	arrayErr := func(code int) *storageapitypes.ResponseStatus {
		return &storageapitypes.ResponseStatus{ResponseTypeNumeric: 1, ReturnCode: code, Response: "Command failed."}
	}
	cases := []struct {
		name   string
		status *storageapitypes.ResponseStatus
		err    error
		want   codes.Code // codes.OK = nil error
	}{
		{"deleted", &storageapitypes.ResponseStatus{}, nil, codes.OK},
		{"no status", nil, nil, codes.OK},
		{"already gone", arrayErr(storageapitypes.SnapshotNotFoundErrorCode), nil, codes.OK},
		{"already gone with err", arrayErr(storageapitypes.SnapshotNotFoundErrorCode), errors.New("x"), codes.OK},
		{"refused", arrayErr(-10183), nil, codes.Internal},
		{"transport error", nil, errors.New("connection refused"), codes.Unknown},
	}
	for _, c := range cases {
		got := status.Code(deleteSnapshotError("0f0f0f0f0f0f0f0f0f0f0f0f", c.status, c.err))
		if got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
