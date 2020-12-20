package main

import (
	"fmt"
	"mydocker/cgroups/subsystems"
	"mydocker/container"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			mydocker run -ti [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container command")
		}
		var cmdArray []string
		//命令行参数放进cmdArray
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		//解析tty
		tty := context.Bool("ti")
		//解析detach
		detach := context.Bool("d")
		if tty && detach {
			return fmt.Errorf("ti and d paramter can not both provided")
		}
		//解析资源限制
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CPUSet:      context.String("cpuset"),
			CPUShare:    context.String("cpushare"),
		}

		//把volume参数传给Run函数
		volume := context.String("v")
		//将取到的容器名称传递下去，如果没有取到值为空
		containerName := context.String("name")
		//开始运行
		Run(tty, cmdArray, resConf, volume, containerName)
		return nil
	},
}

//这里，定义了initCommand 的具体操作，此操作为内部方法，禁止外部调用
var initCommand = cli.Command{
	Name: "init",
	Usage: `Init container process run user’s process in container.
	Do not call it outside`,

	/*
		1. 获取传递过来的 command 参数
		2. 执行容器初始化操作
	*/
	Action: func(context *cli.Context) error {
		log.Infof("init come on")
		err := container.RunContainerInitProcess()
		return err
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container name")
		}
		imageName := context.Args().Get(0)
		commitContainer(imageName)
		return nil
	},
}

func commitContainer(imageName string) {
	mntURL := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	log.Infof("%s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-c", mntURL, "·").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error", mntURL, err)
	}
}
