package systeminfo

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"time"
	"translate-server/structs"
)

func GetSystemInfo() structs.SystemInfo {
	var sysInfo structs.SystemInfo
	sysInfo.CpuPercent = getCpuPercent()
	sysInfo.DiskPercent = getDiskPercent()
	sysInfo.MemPercent = getMemPercent()
	return sysInfo
}


func getCpuPercent() float64 {
	percent, _:= cpu.Percent(time.Second, false)
	return percent[0]
}
func getMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}
func getDiskPercent() float64 {
	parts, _ := disk.Partitions(true)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	return diskInfo.UsedPercent
}

