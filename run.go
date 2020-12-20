package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"mydocker/cgroups"
	"mydocker/cgroups/subsystems"
	"mydocker/container"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

/*
这里的Start方法是真正开始前面创建好的command 的调用，它首先会clone出来namespace隔离的
进程，然后在子进程中，调用/proc/self/exe，也就是调用自己，发送 init参数，
调用我们写的init方法，去初始化容器的一些资源。
*/

//Run ...
func Run(tty bool, cmdArray []string, res *subsystems.ResourceConfig, volume string, containerName string) {
	parent, writePipe := container.NewParentProcess(tty, volume)
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	//记录容器信息
	containerName, err := recordContainerInfo(parent.Process.Pid, cmdArray, containerName)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
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

	if tty {
		parent.Wait()

		cmdUmount := exec.Command("umount", "proc", "tmpfs")
		cmdUmount.Run()

		mntURL := "/root/mnt"
		rootURL := "/root"
		container.DeleteWorkSpace(rootURL, mntURL, volume)
	}
}

//sendInitCommand 向子进程发送初始化命令
func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

/*
这里以时间戳为种子 每次生成 10 以内的数字作为 letter tes 数组的下标，最后拼
接生成整个容器的 ID
*/
func randStringBytes(n int) string {
	letterBytes := "0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func recordContainerInfo(containerPID int, commandArray []string, containerName string) (string, error) {
	id := randStringBytes(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	if containerName == "" {
		containerName = id

	}
	containerInfo := &container.ContainerInfo{
		ID:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Name:        containerName,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return "", err

	}
	jsonStr := string(jsonBytes)

	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirURL, 0622); err != nil {
		log.Errorf("Mkdir error %s error %v", dirURL, err)
		return "", err
	}
	fileName := dirURL + "/" + container.ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("Create file %s error %v", fileName, err)
		return "", err
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string error %v", err)
		return "", err
	}

	return containerName, nil
}

func deleteContainerInfo(containerID string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerID)
	if err := os.RemoveAll(dirURL); err != nil {
		log.Errorf("Remove dir %s error %v", dirURL, err)
	}
}
