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
	"path/filepath"
	"sync"
	"translate-server/config"
	"translate-server/constant"
	"translate-server/datamodels"
	"translate-server/structs"
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
	netId string
}

const ContainerPrefix = "trans_"
const PrivateNetworkName = "trans_network"


func (o *Operator) StartDockers() error {
	service := datamodels.NewActivationModel()
	_, state := service.ParseKeystoreFile()
	compList, err := config.GetInstance().GetComponentList(false)
	if err != nil {
		return err
	}
	o.percent = 0
	totalProcedure := len(compList)
	everyProcedure := 100 / totalProcedure
	for _,v := range compList {
		if state == constant.HttpSuccess || v.DefaultRun {
			err := o.LoadImage(v)
			if err != nil {
				o.status = ErrorStatus
				return err
			}
			o.percent += Percent(everyProcedure / 2)
			id, err := o.StartContainer(v)
			if err != nil {
				o.status = ErrorStatus
				return err
			}
			// 不想大改了，所以就硬编码吧
			if v.ImageName == "redis" || v.ImageName == "mysql" ||v.ImageName == "core" ||v.ImageName == "plugins"  {
				err = o.JoinPrivateNetwork(id)
				if err != nil {
					return err
				}
			}
			o.percent += Percent(everyProcedure / 2)
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
	compList, err := config.GetInstance().GetComponentList(false)
	if err != nil {
		return false, err
	}
	for _,v := range compList {
		running, err := o.isContainerRunning(v.ImageName, v.ImageVersion)
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
func (o *Operator) LoadImage(img structs.ComponentInfo) error {
	b, err := o.existImage(img.ImageName, img.ImageVersion)
	if err != nil {
		return err
	}
	if b {
		return nil
	}
	imgFilePathname := fmt.Sprintf("./components/%s/%s/%s", img.ImageName, img.ImageVersion, img.FileName)
	f, err := os.Open(imgFilePathname)
	if err != nil {
		return err
	}
	_, err = o.cli.ImageLoad(context.Background(), f, true)
	if err != nil {
		return err
	}
	return nil
}


// RemoveContainer 移除容器
func (o *Operator) RemoveContainer(imageName string, imageVersion string) error {
	containers, err := o.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return err
	}
	for _, v := range containers {
		if v.Image == imageName+":"+imageVersion {
			err = o.cli.ContainerStop(context.Background(), v.ID, nil)
			if err != nil {
				return err
			}
			err = o.cli.ContainerRemove(context.Background(), v.ID, types.ContainerRemoveOptions{})
			if err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}
// RemoveImage 移除镜像
func (o *Operator) RemoveImage(imageName string, imageVersion string) error {
	images, err := o.cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return err
	}
	for _, v := range images {
		s := v.RepoTags[0]
		if s == imageName + ":" + imageVersion {
			_, err = o.cli.ImageRemove(context.Background(), v.ID, types.ImageRemoveOptions{})
			if err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

// StartContainer 启动容器 ,如果是 部署在linux下，那么当启动web镜像（nginx）的时候，需要添加--add-host=host.docker.internal:host-gateway参数
func (o *Operator) StartContainer(img structs.ComponentInfo) (string, error) {
	hasContainer, id, err := o.hasContainer(img.ImageName, img.ImageVersion)
	if err != nil {
		return "", err
	}
	if hasContainer {
		running, err := o.isContainerRunning(img.ImageName, img.ImageVersion)
		if err != nil {
			return "", err
		}
		if !running {
			err = o.cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
			if err != nil {
				return "", err
			}
		}
		name, err := o.getContainerNameById(id)
		if err != nil {
			return "", err
		}
		if name !=  ContainerPrefix + img.ImageName{
			err = o.cli.ContainerRename(context.Background(), id,  ContainerPrefix + img.ImageName)
			if err != nil {
				return "", err
			}
		}
		return id, nil
	} else {
		config := &container.Config{
			Image: img.ImageName + ":" + img.ImageVersion,
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
		// web存在于docker中，需要访问主机上api接口
		if img.ImageName == "web" {
			//hostConfig.ExtraHosts = []string{"--add-host=host.docker.internal:host-gateway"}
			hostConfig.ExtraHosts = []string{"host.docker.internal:host-gateway"}
			hostConfig.NetworkMode = "host"
		}
		if img.ImageName == "mysql" {
			mysqlPasswdEnv := fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", structs.MysqlPassword)
			config.Env = []string{mysqlPasswdEnv}
		//	挂载卷本地目录
			dataDir, err := filepath.Abs("./mysql_db/db")
			if err != nil {
				return "", err
			}
			dataConfigDir, err := filepath.Abs("./mysql_db/conf.d")
			if err != nil {
				return "", err
			}
			// 容器内部目录
			containerDataDir := "/var/lib/mysql"
			containerConfigDir := "/etc/mysql/conf.d"
			config.Volumes = map[string]struct{}{
				containerDataDir: {},
				containerConfigDir: {},
			}
			b := fmt.Sprintf("%s:%s", dataDir, containerDataDir)
			b2 := fmt.Sprintf("%s:%s", dataConfigDir, containerConfigDir)
			// 将mongodb中的数据挂载到本地
			hostConfig.Binds = []string {b,b2}
		}
		if img.ImageName == "redis" {
			dataDir, err := filepath.Abs("./redis_db/db")
			if err != nil {
				return "", err
			}
			dataConfigDir, err := filepath.Abs("./redis_db/conf.d")
			if err != nil {
				return "", err
			}
			containerDataDir := "/data"
			containerConfigDir := "/usr/local/etc/redis"
			config.Volumes = map[string]struct{}{
				containerDataDir: {},
				containerConfigDir: {},
			}
			b := fmt.Sprintf("%s:%s", dataDir, containerDataDir)
			b2 := fmt.Sprintf("%s:%s", dataConfigDir, containerConfigDir)
			hostConfig.Binds = []string {b, b2}
		}
		if img.ImageName == "plugins" {
			// 先进行路径的映射，以保证容器可以访问到主机的磁盘文件
			dataDir, err := filepath.Abs("./data")
			if err != nil {
				return "", err
			}
			containerDataDir := dataDir
			config.Volumes = map[string]struct{}{
				containerDataDir: {},
			}
			b := fmt.Sprintf("%s:%s", dataDir, containerDataDir)
			hostConfig.Binds = []string {b}
		}
		create, err := o.cli.ContainerCreate(context.Background(), config, hostConfig, &network.NetworkingConfig{}, nil, ContainerPrefix + img.ImageName)
		if err != nil {
			return "", err
		}
		err = o.cli.ContainerStart(context.Background(), create.ID, types.ContainerStartOptions{})
		if err != nil {
			return "", err
		}
		return create.ID, nil
	}
}

func (o *Operator) getContainerNameById(id string) (string, error) {
	containers, err := o.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return "", err
	}
	for _, v := range containers {
		if v.ID == id {
			return v.Names[0][1:], nil
		}
	}
	return "", nil
}

// ExistImage 镜像是否存在
func (o *Operator) existImage(imageName string, imageTag string) (bool, error) {
	images, err := o.cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return false, err
	}
	for _, image := range images {
		s := image.RepoTags[0]
		if s == imageName + ":" + imageTag {
			return true, nil
		}
	}
	return false, nil
}

// HasContainer 是否存在某个容器 容器的名称默认不指定的时候就是随机的，所以通过遍历ContainerList获取的containers中的每一个容器的镜像名称进行判断即可，【镜像生成容器】
func (o *Operator) hasContainer(imageName string, imageTag string) (bool, string, error) {
	containers, err := o.cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return false, "", err
	}
	for _, v := range containers {
		if v.Image == imageName + ":" + imageTag {
			return true, v.ID, nil
		}
	}
	return false, "", nil
}

// IsContainerRunning 某个容器是否正在运行
func (o *Operator) isContainerRunning(imageName string, imageTag string) (bool, error) {
	containers, err := o.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return false, err
	}
	for _, v := range containers {
		if v.Image == imageName + ":" + imageTag {
			return true, nil
		}
	}
	return false, nil
}

// CreatePrivateNetwork 创建一个翻译系统的私有网络 名称由PrivateNetworkName变量设定
func (o* Operator) CreatePrivateNetwork() error {
	list, err := o.cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return err
	}
	for _, v := range list {
		if  v.Name == PrivateNetworkName {
			o.netId = v.ID
			return nil
		}
	}
	resp, err := o.cli.NetworkCreate(context.Background(), PrivateNetworkName, types.NetworkCreate{})
	if err != nil {
		return err
	}
	o.netId = resp.ID
	return nil
}

func (o Operator) JoinPrivateNetwork(containerId string) error {
	inspect, err := o.cli.NetworkInspect(context.Background(), o.netId, types.NetworkInspectOptions{})
	if err != nil {
		return err
	}
	for k, _ := range inspect.Containers{
		if k == containerId {
			return nil
		}
	}
	return o.cli.NetworkConnect(context.Background(), o.netId, containerId, nil)
}