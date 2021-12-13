package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"os"
	"strings"
	"sync"
)

type ContainerInfo struct {
	ImageName     string
	ContainerName string
	LoadFilePath  string
	Config        *container.Config
	HostConfig    *container.HostConfig
}

var ContainerList []ContainerInfo

func init()  {
	// tika 配置
	tika := ContainerInfo{
		ImageName:     "tika",
		ContainerName: "tika",
		LoadFilePath:  "./tika.tar",
	}
	config := &container.Config{
		Image: tika.ImageName,
		ExposedPorts: nat.PortSet{
			"9998/tcp": {},
		}}
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"9998/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "9998",
				},
			},
		},
	}
	tika.Config = config
	tika.HostConfig = hostConfig

	ContainerList = append(ContainerList, tika)
}

func init()  {
	// tika 配置
	translate := ContainerInfo{
		ImageName:     "translate",
		ContainerName: "translate",
		LoadFilePath:  "./translate.tar",
	}
	config := &container.Config{
		Image: translate.ImageName,
		ExposedPorts: nat.PortSet{
			"5000/tcp": {},
		}}
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"5000/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "5000",
				},
			},
		},
	}
	translate.Config = config
	translate.HostConfig = hostConfig

	ContainerList = append(ContainerList, translate)
}

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
		instance.status = Normal
	})
	return instance
}
type Status int

const (
	Initializing Status = iota // 激活后第一次的初始化
	Normal
)

type Operator struct {
	cli *client.Client
	status Status   // 是否正在初始化
}

func (o *Operator) StartDockers() error {
	for _,v := range ContainerList {
		err := o.loadImage(v)
		if err != nil {
			return err
		}
		err = o.startContainer(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Operator) SetStatus(status Status)  {
	o.status = status
}
func (o *Operator) GetStatus() Status {
	return o.status
}

func (o *Operator) IsALlRunningStatus() (bool, error) {
	for _,v := range ContainerList {
		running, err := o.isContainerRunning(v.ContainerName)
		if err != nil {
			return false, err
		}
		if !running {
			return false, nil
		}
	}
	return true,nil
}

// LoadImage 从文件加载镜像
func (o *Operator) loadImage(info ContainerInfo) error {
	b, err := o.existImage(info)
	if err != nil {
		return err
	}
	if b {
		return nil
	}
	f, err := os.Open(info.LoadFilePath)
	if err != nil {
		return err
	}
	_, err = o.cli.ImageLoad(context.Background(), f, true)
	if err != nil {
		return err
	}
	return nil
}

func ReadBigFile(fileName string, handle func([]byte)) error {
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println("can't opened this file")
		return err
	}
	defer f.Close()
	s := make([]byte, 4096)
	for {
		switch nr, err := f.Read(s[:]); true {
		case nr < 0:
			fmt.Fprintf(os.Stderr, "cat: error reading: %s\n", err.Error())
			os.Exit(1)
		case nr == 0: // EOF
			return nil
		case nr > 0:
			handle(s[0:nr])
		}
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
func (o *Operator) startContainer(info ContainerInfo) error {
	hasContainer, id, err := o.hasContainer(info.ContainerName)
	if err != nil {
		return err
	}
	if hasContainer {
		running, err := o.isContainerRunning(info.ContainerName)
		if err != nil {
			return err
		}
		if running {
			return o.cli.ContainerRestart(context.Background(), id, nil)
		} else {
			return o.cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
		}
	} else {
		create, err := o.cli.ContainerCreate(context.Background(), info.Config, info.HostConfig, &network.NetworkingConfig{}, nil, "")
		if err != nil {
			return err
		}
		return o.cli.ContainerStart(context.Background(), create.ID, types.ContainerStartOptions{})
	}
}

// ExistImage 镜像是否存在
func (o *Operator) existImage(info ContainerInfo) (bool, error) {
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

// HasContainer 是否存在某个容器
func (o *Operator) hasContainer(name string) (bool, string, error) {
	containers, err := o.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return false, "", err
	}
	for _, v := range containers {
		if strings.Contains(v.Image, name) {
			return true, v.ID, nil
		}
	}
	return false, "", nil
}

// IsContainerRunning 某个容器是否正在运行
func (o *Operator) isContainerRunning(name string) (bool, error) {
	containers, err := o.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return false, err
	}
	for _, v := range containers {
		if strings.Contains(v.Image, name) {
			return true, nil
		}
	}
	return false, nil
}