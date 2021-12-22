package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"os"
	"strings"
	"sync"
	"translate-server/config"
	"translate-server/datamodels"
	"translate-server/services"
)


var instance *Operator
var once sync.Once

func GetInstance() *Operator {
	once.Do(func() {
		instance = &Operator{}
		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			panic(err)
		}
		instance.cli = cli
		instance.percent = 0
		instance.status = NormalStatus
	})
	return instance
}
type Status int
type Percent int

const (
	InitializingStatus Status = iota // 激活后第一次的初始化
	RepairingStatus
	NormalStatus
	ErrorStatus
)

type Operator struct {
	cli *client.Client
	status Status   // 是否正在初始化
	percent Percent
}

func (o *Operator) StartDockers() error {
	service := services.NewActivationService()
	_, state := service.ParseKeystoreFile()
	systemConfig, err := config.GetInstance().ParseSystemConfigFile(false)
	if err != nil {
		return err
	}
	o.percent = 0
	for _,v := range systemConfig.ComponentList {
		if state == datamodels.HttpSuccess || v.DefaultRun {
			err := o.loadImage(v)
			if err != nil {
				o.status = ErrorStatus
				return err
			}
			o.percent += 15
			err = o.startContainer(v)
			if err != nil {
				o.status = ErrorStatus
				return err
			}
			o.percent += 10
		}
	}
	o.percent = 100
	return nil
}

func (o *Operator) SetPercent(percent Percent)  {
	o.percent = percent
}
func (o *Operator) GetPercent() Percent {
	return o.percent
}

func (o *Operator) SetStatus(status Status)  {
	o.status = status
}
func (o *Operator) GetStatus() Status {
	return o.status
}

func (o *Operator) IsALlRunningStatus() (bool, error) {
	systemConfig, err := config.GetInstance().ParseSystemConfigFile(false)
	if err != nil {
		return false, err
	}
	for _,v := range systemConfig.ComponentList {
		running, err := o.isContainerRunning(v.ImageName)
		if err != nil {
			return false,  err
		}
		if !running {
			return false, nil
		}
	}
	return true, nil
}

// LoadImage 从文件加载镜像
func (o *Operator) loadImage(img datamodels.ComponentInfo) error {
	b, err := o.existImage(img)
	if err != nil {
		return err
	}
	if b {
		return nil
	}
	f, err := os.Open(img.FileName)
	if err != nil {
		return err
	}
	_, err = o.cli.ImageLoad(context.Background(), f, true)
	if err != nil {
		return err
	}
	return nil
}


// RemoveAllContainer 移除所有容器 包括运行的和没有运行的
func (o *Operator) RemoveAllContainer() error {
	containers, err := o.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return err
	}
	for _, v := range containers {
		err = o.cli.ContainerStop(context.Background(), v.ID, nil)
		if err != nil {
			return err
		}
		err = o.cli.ContainerRemove(context.Background(), v.ID, types.ContainerRemoveOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Operator) RemoveImage(id string) error {
	_, err := o.cli.ImageRemove(context.Background(), id, types.ImageRemoveOptions{})
	if err != nil {
		return err
	}
	return nil
}

// StartContainer 启动容器
func (o *Operator) startContainer(img datamodels.ComponentInfo) error {
	hasContainer, id, err := o.hasContainer(img.ImageName)
	if err != nil {
		return err
	}
	if hasContainer {
		running, err := o.isContainerRunning(img.ImageName)
		if err != nil {
			return err
		}
		if running {
			//return o.cli.ContainerRestart(context.Background(), id, nil)
			return nil
		} else {
			return o.cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
		}
	} else {
		config := &container.Config{
			Image: img.ImageName,
			ExposedPorts: nat.PortSet{
				nat.Port(img.ExposedPort + "/tcp"): {},
			},
		}
		hostConfig := &container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port(img.ExposedPort + "/tcp"): []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: img.HostPort,
					},
				},
			},
		}
		create, err := o.cli.ContainerCreate(context.Background(), config, hostConfig, &network.NetworkingConfig{}, nil, "")
		if err != nil {
			return err
		}
		return o.cli.ContainerStart(context.Background(), create.ID, types.ContainerStartOptions{})
	}
}

// ExistImage 镜像是否存在
func (o *Operator) existImage(info datamodels.ComponentInfo) (bool, error) {
	images, err := o.cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return false, err
	}
	for _, image := range images {
		s := image.RepoTags[0]
		arrays := strings.Split(s, ":")
		if strings.Contains(arrays[0], info.ImageName) {
			return true, nil
		}
	}
	return false, nil
}

// HasContainer 是否存在某个容器 容器的名称默认不指定的时候就是随机的，所以通过遍历ContainerList获取的containers中的每一个容器的镜像名称进行判断即可，【镜像生成容器】
func (o *Operator) hasContainer(imageName string) (bool, string, error) {
	containers, err := o.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return false, "", err
	}
	for _, v := range containers {
		if v.Image == imageName {
			return true, v.ID, nil
		}
	}
	return false, "", nil
}

// IsContainerRunning 某个容器是否正在运行
func (o *Operator) isContainerRunning(imageName string) (bool, error) {
	containers, err := o.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return false, err
	}
	for _, v := range containers {
		if v.Image == imageName {
			return true, nil
		}
	}
	return false, nil
}