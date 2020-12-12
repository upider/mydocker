package container

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

/*
这里是父进程，也就是当前进程执行的内容，根据上 章介绍的内容，应该比较容易明白。
1. 这里的/proc/self/exe 调用中，/proc/self/指的是当前运行进程自己的环境，exec其实就是自己
调用了自己，使用这种方式对创建出来的进程进行初始化
2. 后面的 args 是参数，其中 init 是传递给本进程的第 个参数，在本例中，其实就是会去调用 initCornmand
去初始化进程的 些环境和资源
3. 下面的 clone 参数就是去 fork 出来一个新进程，并且使用了 name space 隔离新创建的进程和外部环境
4. 如果用户指定了－ti 参数，就需要把当前进程的输入输出导入到标准输入输出上
*/

// NewParentProcess ...
func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}

	rootURL := "/root"
	mntURL := "/root/mnt"
	NewWorkSpace(rootURL, mntURL, volume)
	cmd.Dir = mntURL

	return cmd, writePipe
}

//NewPipe 创建管道
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

//NewWorkSpace Create a overlay filesystem as container root workspace
func NewWorkSpace(rootURL string, mntURL string, volume string) {
	CreateReadOnlyLayer(rootURL)
	CreateUpper(rootURL)
	CreateWork(rootURL + "work")
	CreateMountPoint(mntURL)
	if volume != "" {
		MountVolume(rootURL, mntURL, volume)
	}
}

//解析volume字符串
func volumeURLExtract(volume string) []string {
	var volumeURLs []string
	volumeURLs = strings.Split(volume, ":")
	return volumeURLs
}

/*
挂载数据卷的过程如下。
l. 首先，读取宿主机文件目录 URL ，创建宿主机文件目录(/root/${parentUrl)。
2. 然后，读取容器挂载点 URL，在容器文件系统里创建挂载点(/root/mnt/${containerUrl})
3. 最后，把宿主机文件目录挂载到容器挂载点。这样启动容器的过程，对数据卷的处理也就完成了。
*/

//MountVolume ...
func MountVolume(rootURL string, mntURL string, volume string) {
	//创建宿主机文件目录
	volumeURLs := volumeURLExtract(volume)
	parentURL := volumeURLs[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		log.Infof("Mkdir parent dir %s error. %v", parentURL, err)
	}
	//在容器文件系统里创建挂载点
	containerURL := volumeURLs[1]
	containerVolumeURL := mntURL + containerURL
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Infof("Mkdir container dir %s error. %v", containerVolumeURL, err)
	}
	cmd := exec.Command("mount", "--bind", volumeURLs[0], containerVolumeURL)
	log.Info("mount volume")
	log.Infof("%v", cmd.Args)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Mount volume failed. %v", err)
	}
}

//CreateReadOnlyLayer `将busybox.tar解压到busybox目录下，作为容器的只读层`
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("Untar dir %s error. %v", busyboxURL, err)
		}
	}
}

//CreateWork 创建了一个名为work的文件夹作为容器唯一的可写层
func CreateWork(rootURL string) {
	workURL := rootURL + "/work"
	exist, err := PathExists(workURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", workURL, err)
	}
	if exist == false {
		if err := os.Mkdir(workURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", workURL, err)
		}
	}
}

//CreateUpper 创建了一个名为upper的文件夹作为容器唯一的可写层
func CreateUpper(rootURL string) {
	writeURL := rootURL + "/upper"
	exist, err := PathExists(writeURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", writeURL, err)
	}
	if exist == false {
		if err := os.Mkdir(writeURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", writeURL, err)
		}
	}
}

//CreateMountPoint 创建mnt文件夹作为挂载点
func CreateMountPoint(mntURL string) {
	exist, err := PathExists(mntURL)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", mntURL, err)
	}
	if exist == false {
		if err := os.Mkdir(mntURL, 0777); err != nil {
			log.Errorf("Mkdir dir %s error. %v", mntURL, err)
		}
	}

	mergeDir := "/root/mnt"
	upperDir := "upperdir=/root/upper"
	lowerDir := "lowerdir=/root/busybox," + upperDir + ",workdir=/root/work"
	//dirs := lowerDir + " " + mergeDir
	log.Info("mount overlay")
	//把宿主机文件目录挂载到容器挂载点
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", lowerDir, mergeDir)
	log.Infof("%v", cmd.Args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Mount overlay failed. %v", err)
	}
}

//PathExists 判断文件路径是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

/*
删除容器文件系统的过程如下。
1. 只有在volume不为空，并且使用volumeURLExtract函数解volume字符返回的字符数
组长度为 ，数据元素均不为空的DeleteMountPointWithVolume函数来处理。
2. 其余的情况然使用前面的DeleteMountPoint函数
*/

//DeleteWorkSpace Delete the overlayFS filesystem while container exit
func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	if volume != "" {
		volumeURLs := volumeURLExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(rootURL, mntURL, volumeURLs)
		} else {
			DeleteMountPoint(rootURL, mntURL)
		}
	} else {
		DeleteMountPoint(rootURL, mntURL)
	}
	DeleteUpper(rootURL)
	DeleteWork(rootURL + "/work")
}

// DeleteMountPoint ...
func DeleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("DeleteMountPoint Error: %v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
	}
}

/*
leteMountPointWithVolum 函数的处理逻辑如下
1. 首先，卸载volume挂载点的文件系统/root/mnt/${containerURL}保证整个容器的挂载点没有被使用。
2. 然后，再卸载整个容器文件系统挂载点/root/mnt。
3. 最后，删除容器文件系统挂载点。整个容器退出过程中的件系统处理就结束了。
*/

//DeleteMountPointWithVolume  ...
func DeleteMountPointWithVolume(rootURL string, mntURL string, volumeURLs []string) {
	//卸载容器里volume挂载点的文件系统
	containerURL := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerURL)
	log.Infof("%v", cmd.Args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount volume failed. %v", err)
	}
	//卸载整个容器文件系统的挂载点
	cmd = exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount mountpoint failed．%v", err)
	}
	//删除容器文件系统挂载点
	if err := os.RemoveAll(mntURL); err != nil {
		log.Infof("Remove mountpoint dir %s  error %v", mntURL, err)
	}
}

//DeleteWork ...
func DeleteWork(rootURL string) {
	workURL := rootURL
	if err := os.RemoveAll(workURL); err != nil {
		log.Errorf("Remove dir %s error %v", workURL, err)
	}
}

//DeleteUpper ...
func DeleteUpper(rootURL string) {
	writeURL := rootURL + "/upper"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
	}
}
