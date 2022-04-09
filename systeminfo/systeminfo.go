package systeminfo

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"time"
	"translate-server/structs"
)


func GetSystemInfo() structs.SystemInfo {
	var sysInfo structs.SystemInfo
	sysInfo.CpuPercent = getCpuPercent()
	sysInfo.IoReadBytes, sysInfo.IoWriteBytes = getDiskIo()
	sysInfo.IoReadBytesStr = formatSize(sysInfo.IoReadBytes)
	sysInfo.IoWriteBytesStr = formatSize(sysInfo.IoWriteBytes)
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
// 字节的单位转换 保留两位小数
func formatSize(fileSize uint64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}
func getDiskIoOnce()  (uint64,uint64) {
	counters, err := disk.IOCounters()
	if err != nil {
		fmt.Println(err)
		return 0, 0
	}
	var ReadBytes uint64
	var WriteBytes uint64
	for _, c := range counters {
		ReadBytes += c.ReadBytes
		WriteBytes += c.WriteBytes
	}
	return ReadBytes, WriteBytes
}

func getDiskIo() (uint64, uint64) {
	preReadBytes, preWriteBytes := getDiskIoOnce()
	time.Sleep(time.Second)
	currentReadBytes, currentWriteBytes := getDiskIoOnce()
	readSpeed := currentReadBytes - preReadBytes
	writeSpeed := currentWriteBytes - preWriteBytes
	return readSpeed, writeSpeed
}

