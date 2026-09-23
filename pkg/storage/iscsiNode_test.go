package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSysfsWwid(t *testing.T) {
	for in, want := range map[string]string{
		"600C0FF000FB6CA93805B36A01000000":  "naa.600c0ff000fb6ca93805b36a01000000",
		"3600c0ff000fb6ca93805b36a01000000": "naa.600c0ff000fb6ca93805b36a01000000",
	} {
		if got := sysfsWwid(in); got != want {
			t.Errorf("sysfsWwid(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestScsiDevicesByWwn(t *testing.T) {
	root := t.TempDir()
	old := sysBlockPath
	sysBlockPath = root
	t.Cleanup(func() { sysBlockPath = old })

	for dev, wwid := range map[string]string{
		"sda": "naa.6f4ee0806f0ce2002dbfe9acd658df80", // local disk
		"sdb": "naa.600c0ff000fb6ca93805b36a01000000\n",
		"sdc": "naa.600c0ff000fb6ca93805b36a01000000",
		"sdd": "naa.600c0ff000fb6ca945fbb26a01000000", // another LUN
	} {
		d := filepath.Join(root, dev, "device")
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "wwid"), []byte(wwid), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// dm devices have no device/wwid; must be ignored, not crash
	if err := os.MkdirAll(filepath.Join(root, "dm-31", "slaves"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := scsiDevicesByWwn("600c0ff000fb6ca93805b36a01000000")
	if len(got) != 2 || got[0] != "sdb" || got[1] != "sdc" {
		t.Errorf("scsiDevicesByWwn = %v, want [sdb sdc]", got)
	}
	if got := scsiDevicesByWwn(""); got != nil {
		t.Errorf("empty wwn must match nothing, got %v", got)
	}
	if !multipathDeviceIsOrphan("/dev/dm-31") {
		t.Error("dm-31 with empty slaves dir must be orphan")
	}
}
