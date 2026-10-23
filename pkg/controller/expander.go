package controller

import (
	"context"
	"fmt"

	"github.com/Seagate/seagate-exos-x-csi/pkg/common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

// ControllerExpandVolume expands a volume to the given new size
func (controller *Controller) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	volumeName, _ := common.VolumeIdGetName(req.GetVolumeId())
	if volumeName == "" {
		return nil, status.Error(codes.InvalidArgument, "cannot expand a volume with an empty ID")
	}
	klog.Infof("expanding volume %q", volumeName)

	newSize := req.GetCapacityRange().GetRequiredBytes()
	if newSize == 0 {
		newSize = req.GetCapacityRange().GetLimitBytes()
	}
	klog.V(2).Infof("requested size: %d bytes", newSize)

	if err := controller.expandVolumeTo(volumeName, newSize); err != nil {
		return nil, err
	}

	klog.Infof("volume %q successfully expanded", volumeName)

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         newSize,
		NodeExpansionRequired: true,
	}, nil
}

// expandVolumeTo grows volumeName to at least newSize bytes. A volume already
// that large is left alone, so a repeated call is a no-op (CSI requires
// ControllerExpandVolume and CreateVolume to be idempotent).
func (controller *Controller) expandVolumeTo(volumeName string, newSize int64) error {
	response, _, err := controller.client.ShowVolumes(volumeName)
	if err != nil {
		return err
	} else if len(response) == 0 {
		return fmt.Errorf("volume %q not found", volumeName)
	} else if response[0].SizeNumeric == 0 {
		return fmt.Errorf("could not get current volume size, thus volume expansion is not possible")
	} else if response[0].Blocks == 0 {
		return fmt.Errorf("could not parse volume size: %v", err)
	}

	currentSize := response[0].Blocks * response[0].BlockSize
	expansionSize := bytesToGrow(currentSize, newSize)
	klog.V(2).Infof("volume %q current size: %d bytes, expanding by %d bytes", volumeName, currentSize, expansionSize)
	if expansionSize == 0 {
		return nil
	}
	_, err = controller.client.ExpandVolume(volumeName, getSizeStr(expansionSize))
	return err
}

// bytesToGrow is how much a volume of current bytes must grow to hold required.
func bytesToGrow(current, required int64) int64 {
	if required <= current {
		return 0
	}
	return required - current
}
