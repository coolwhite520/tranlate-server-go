package structs

type SystemInfo struct {
	CpuPercent float64 `json:"cpu_percent"`
	MemPercent float64 `json:"mem_percent"`
	DiskPercent float64 `json:"disk_percent"`
}
