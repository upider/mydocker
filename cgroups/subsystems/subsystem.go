package subsystems

// ResourceConfig 用于传递资源限制配直的结构体，包含内存限制，CPU时间片权重CPU核心数
type ResourceConfig struct {
	MemoryLimit string
	CPUShare    string
	CPUSet      string
}

// Subsystem 每个Subsystem可以实现下面的 个接口
//这里将cgroup抽象成了path原因是cgroup在hierarchy的路径，便是虚拟文件系统中的虚拟路径
type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

var (
	// SubsystemsIns 通过不同subsystem初始化实例创建资源限制处理链数组
	SubsystemsIns = []Subsystem{
		&CPUSetSubSystem{},
		&MemorySubSystem{},
		&CPUSubSystem{},
	}
)
