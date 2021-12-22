package datamodels

type DockerImg struct {
	FileName      string `json:"file_name"`
	ImageName     string `json:"image_name"`
	ContainerName string `json:"container_name"`
	ContainerTag  string `json:"container_tag"` // 可以认为是version
	FileMd5       string `json:"file_md5"`
	InternalPort  string `json:"internal_port"` // 服务内部端口
	ExposePort    string `json:"expose_port"` // 对外暴露端口
	DefaultRun    bool   `json:"default_run"`  // 默认是否启动
}
type DockerImgList []DockerImg