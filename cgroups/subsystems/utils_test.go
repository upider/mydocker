package subsystems

import (
	"os"
	"path"
	"testing"
)

func TestFindCgroupMountpoint(t *testing.T) {
	t.Logf("cpu subsystem mount point %v\n", FindCgroupMountPoint("cpu"))
	t.Logf("cpuset subsystem mount point %v\n", FindCgroupMountPoint("cpuset"))
	t.Logf("memory subsystem mount point %v\n", FindCgroupMountPoint("memory"))
}

func TestGetCgroupPath(t *testing.T) {
	cgroupPath := "mydocker-cgroup"
	subsystem := "memory"
	subsysCgroupPath, _ := GetCgroupPath(subsystem, cgroupPath, false)
	//if err != nil {
	//    fmt.Errorf("GetCgroupPath Error: %v", err)
	//}
	cgroupRoot := FindCgroupMountPoint(subsystem)
	t.Logf(path.Join(cgroupRoot, cgroupPath))
	t.Logf("subsysCgroupPath is %s", subsysCgroupPath)

	info, err := os.Stat(path.Join(cgroupRoot, cgroupPath))
	if err == nil {
		t.Log(info)
	} else {
		t.Logf("%s", err)
	}
}
