package main

import (
	"mydocker/cgroups"
	"mydocker/cgroups/subsystems"
	"mydocker/container"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

/*
这里的Start方法是真正开始前面创建好的command 的调用，它首先会clone出来namespace隔离的
进程，然后在子进程中，调用/proc/self/exe，也就是调用自己，发送 init参数，
调用我们写的init方法，去初始化容器的一些资源。
*/

//Run ...
func Run(tty bool, cmdArray []string, res *subsystems.ResourceConfig, volume string) {
	parent, writePipe := container.NewParentProcess(tty, volume)
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	sendInitCommand(cmdArray, writePipe)

	//use mydocker-cgroup as cgroup name
	//创建 cgroup manager，并通过调用set apply设置资源限制并使限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	//设置资源限制
	cgroupManager.Set(res)
	//将容器进程加入到各个subsystem挂载对应的cgroup
	cgroupManager.Apply(parent.Process.Pid)
	//对容器设置完限制之后初始容器

	parent.Wait()

	cmdUmount := exec.Command("umount", "proc", "tmpfs")
	cmdUmount.Run()
	mntURL := "/root/mnt"
	rootURL := "/root"
	container.DeleteWorkSpace(rootURL, mntURL, volume)

	os.Exit(0)
}

//sendInitCommand 向子进程发送初始化命令
func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
