package structs

type SystemInfo struct {
	CpuPercent      float64 `json:"cpu_percent"`
	MemPercent      float64 `json:"mem_percent"`
	IoReadBytes     uint64  `json:"io_read_bytes"`
	IoWriteBytes    uint64  `json:"io_write_bytes"`
	IoReadBytesStr  string  `json:"io_read_bytes_str"`
	IoWriteBytesStr string  `json:"io_write_bytes_str"`
}
