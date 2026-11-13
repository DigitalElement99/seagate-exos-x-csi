package controller

import (
	"errors"
	"testing"

	storageapitypes "github.com/Seagate/seagate-exos-x-api-go/v2/pkg/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The array answers a refused delete with HTTP 200 and an Error status, which
// api-go returns as err == nil.
func TestDeleteVolumeError(t *testing.T) {
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
		{"already gone", arrayErr(storageapitypes.VolumeNotFoundErrorCode), nil, codes.OK},
		{"has snapshots", arrayErr(storageapitypes.VolumeHasSnapshot), nil, codes.FailedPrecondition},
		{"other array error", arrayErr(-12345), nil, codes.Internal},
		{"transport error", nil, errors.New("connection refused"), codes.Unknown},
		{"has snapshots with err", arrayErr(storageapitypes.VolumeHasSnapshot), errors.New("x"), codes.FailedPrecondition},
	}
	for _, c := range cases {
		got := status.Code(deleteVolumeError("k3s_test", c.status, c.err))
		if got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
