package datamodels

const SystemVersion = "5.3.16"

type ComponentInfo struct {
	FileName      string `json:"file_name"`
	ImageName     string `json:"image_name"`
	ImageVersion  string `json:"image_version"` // 镜像版本
	FileMd5       string `json:"file_md5"`
	ExposedPort   string `json:"expose_port"` // 服务暴露的端口
	HostPort      string `json:"host_port"`        // 映射的主机端口
	DefaultRun    bool   `json:"default_run"` // 默认是否启动
}
type ComponentList []ComponentInfo
