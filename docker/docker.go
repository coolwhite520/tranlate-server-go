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
		ImageName:     "apache/tika",
		ContainerName: "apache/tika",
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
	})
	return instance
}

type Operator struct {
	cli *client.Client
}

func (d *Operator) StartDockers() error {
	for _,v := range ContainerList {
		err := d.loadImage(v)
		if err != nil {
			return err
		}
		err = d.startContainer(v)
		if err != nil {
			return err
		}
	}
	return nil
}
// LoadImage 从文件加载镜像
func (d *Operator) loadImage(info ContainerInfo) error {
	b, err := d.existImage(info)
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
	_, err = d.cli.ImageLoad(context.Background(), f, true)
	if err != nil {
		return err
	}
	return nil
}

// RemoveAllContainer 移除所有容器 包括运行的和没有运行的
func (d *Operator) RemoveAllContainer() error {
	containers, err := d.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return err
	}
	for _, v := range containers {
		err = d.cli.ContainerStop(context.Background(), v.ID, nil)
		if err != nil {
			return err
		}
		err = d.cli.ContainerRemove(context.Background(), v.ID, types.ContainerRemoveOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// StartContainer 启动容器
func (d *Operator) startContainer(info ContainerInfo) error {
	hasContainer, id, err := d.hasContainer(info.ContainerName)
	if err != nil {
		return err
	}
	if hasContainer {
		running, err := d.isContainerRunning(info.ContainerName, "")
		if err != nil {
			return err
		}
		if running {
			return d.cli.ContainerRestart(context.Background(), id, nil)
		} else {
			return d.cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
		}
	} else {
		create, err := d.cli.ContainerCreate(context.Background(), info.Config, info.HostConfig, &network.NetworkingConfig{}, nil, "")
		if err != nil {
			return err
		}
		return d.cli.ContainerStart(context.Background(), create.ID, types.ContainerStartOptions{})
	}
}

// ExistImage 镜像是否存在
func (d *Operator) existImage(info ContainerInfo) (bool, error) {
	images, err := d.cli.ImageList(context.Background(), types.ImageListOptions{})
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
func (d *Operator) hasContainer(name string) (bool, string, error) {
	containers, err := d.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
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
func (d *Operator) isContainerRunning(name string, id string) (bool, error) {
	containers, err := d.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return false, err
	}
	for _, v := range containers {
		if strings.Contains(v.Image, name) || strings.Contains(v.ID, id) {
			return true, nil
		}
	}
	return false, nil
}