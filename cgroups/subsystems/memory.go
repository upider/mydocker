package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

//MemorySubSystem memory subsystem 的实现
type MemorySubSystem struct{}

//Set 设置cgroup Path对应的cgroup的内存资源限制
func (s *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	//GetCgroupPath 的作用是获取当前 subsystem 在虚拟文件系统中的路径
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.MemoryLimit != "" {
			/*设置这个cgroup的内存限制，即将限制写入到cgroup对应目录的
			memory.limit_in_bytes文件中。
			*/
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"),
				[]byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroup memory fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

//Remove 删除cgroupPath对应的cgroup
func (s *MemorySubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		//删除cgroup便是删除对应的cgroupPath的目录
		return os.Remove(subsysCgroupPath)
	} else {
		return err
	}
}

//Apply 将1个迸程加入到cgroupPath对应的cgroup中
func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		//把进程的PID写到cgroup的虚拟文件系统对应目录下的"task"文件中
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail: %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

//Name 返回cgroup的名字
func (s *MemorySubSystem) Name() string {
	return "memory"
}
