package node

import (
	"bufio"
	"bytes"
	"flutelake/fluteNAS/pkg/model"
	"flutelake/fluteNAS/pkg/module/flog"
	"flutelake/fluteNAS/pkg/module/metricsvm"
	"flutelake/fluteNAS/pkg/util"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DiskUsage struct {
	MountPoint    string
	TotalBytes    uint64
	UsedBytes     uint64
	UsagePercent  float64
	IsSystemDisk  bool
	IsHDD         bool
	Filesystem    string
	Device        string
	SpecMountPath string
}

type NodeMetrics struct {
	CPUUsagePercent  float64
	Load1            float64
	Load5            float64
	Load15           float64
	MemTotalBytes    uint64
	MemUsedBytes     uint64
	MemUsagePercent  float64
	RootTotalBytes   uint64
	RootUsedBytes    uint64
	RootUsagePercent float64
	DataDisks        []DiskUsage
}

type ServiceMetrics struct {
	Installed         bool
	Status            string
	ActiveConnections int64
}

type MonitoringMetrics struct {
	HostIP    string
	Timestamp string
	Node      NodeMetrics
	Samba     ServiceMetrics
	NFS       ServiceMetrics
}

type diskUsageCacheEntry struct {
	usage      DiskUsage
	lastSample time.Time
}

var diskUsageCache = struct {
	mu   sync.Mutex
	data map[string]diskUsageCacheEntry
}{
	data: make(map[string]diskUsageCacheEntry),
}

func CollectSelfMonitoringMetrics() {
	GetMonitoringMetrics("127.0.0.1")
}

func GetMonitoringMetrics(hostIP string) (MonitoringMetrics, error) {
	nodeMetrics, err := collectNodeMetrics(hostIP)
	if err != nil {
		return MonitoringMetrics{}, err
	}

	sambaMetrics, err := collectSambaMetrics(hostIP)
	if err != nil {
		flog.Warnf("collect samba metrics failed on host %s: %v", hostIP, err)
	}

	nfsMetrics, err := collectNFSMetrics(hostIP)
	if err != nil {
		flog.Warnf("collect nfs metrics failed on host %s: %v", hostIP, err)
	}

	metricsvm.UpdateNodeMetrics(
		hostIP,
		nodeMetrics.CPUUsagePercent,
		nodeMetrics.MemUsagePercent,
		nodeMetrics.RootUsagePercent,
		nodeMetrics.Load1,
		nodeMetrics.Load5,
		nodeMetrics.Load15,
	)
	for _, d := range nodeMetrics.DataDisks {
		metricsvm.UpdateDiskUsage(
			hostIP,
			d.MountPoint,
			d.Device,
			d.Filesystem,
			d.IsSystemDisk,
			d.IsHDD,
			d.UsagePercent,
			d.TotalBytes,
			d.UsedBytes,
		)
	}
	metricsvm.UpdateServiceMetrics(
		hostIP,
		"samba",
		sambaMetrics.Installed,
		sambaMetrics.Status,
		sambaMetrics.ActiveConnections,
	)
	metricsvm.UpdateServiceMetrics(
		hostIP,
		"nfs",
		nfsMetrics.Installed,
		nfsMetrics.Status,
		nfsMetrics.ActiveConnections,
	)

	return MonitoringMetrics{
		HostIP:    hostIP,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Node:      nodeMetrics,
		Samba:     sambaMetrics,
		NFS:       nfsMetrics,
	}, nil
}

func collectNodeMetrics(hostIP string) (NodeMetrics, error) {
	cpuUsage, err := collectCPUUsage(hostIP)
	if err != nil {
		return NodeMetrics{}, err
	}

	load1, load5, load15, err := collectLoadAverage(hostIP)
	if err != nil {
		return NodeMetrics{}, err
	}

	memTotal, memUsed, memUsagePercent, err := collectMemoryUsage(hostIP)
	if err != nil {
		return NodeMetrics{}, err
	}

	rootTotal, rootUsed, rootUsagePercent, err := collectRootDiskUsage(hostIP)
	if err != nil {
		return NodeMetrics{}, err
	}

	dataDisks, err := collectDataDiskUsages(hostIP)
	if err != nil {
		return NodeMetrics{}, err
	}

	return NodeMetrics{
		CPUUsagePercent:  cpuUsage,
		Load1:            load1,
		Load5:            load5,
		Load15:           load15,
		MemTotalBytes:    memTotal,
		MemUsedBytes:     memUsed,
		MemUsagePercent:  memUsagePercent,
		RootTotalBytes:   rootTotal,
		RootUsedBytes:    rootUsed,
		RootUsagePercent: rootUsagePercent,
		DataDisks:        dataDisks,
	}, nil
}

func collectCPUUsage(hostIP string) (float64, error) {
	exec := NewExec().SetHost(hostIP)
	defer exec.Close()

	output, err := exec.Command("head -n 1 /proc/stat; sleep 0.1; head -n 1 /proc/stat")
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	var idle1, total1, idle2, total2 uint64
	if scanner.Scan() {
		line := scanner.Text()
		idle1, total1, err = parseCPUSample(line)
		if err != nil {
			return 0, err
		}
	}
	if scanner.Scan() {
		line := scanner.Text()
		idle2, total2, err = parseCPUSample(line)
		if err != nil {
			return 0, err
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	if total2 <= total1 {
		return 0, nil
	}

	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)
	if totalDelta == 0 {
		return 0, nil
	}

	usage := (1 - idleDelta/totalDelta) * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	return usage, nil
}

func parseCPUSample(line string) (uint64, uint64, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return 0, 0, fmt.Errorf("invalid cpu line: %s", line)
	}

	var values []uint64
	for i := 1; i < len(fields); i++ {
		v, err := strconv.ParseUint(fields[i], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		values = append(values, v)
	}

	var total uint64
	for _, v := range values {
		total += v
	}

	idle := values[3]
	return idle, total, nil
}

func collectLoadAverage(hostIP string) (float64, float64, float64, error) {
	exec := NewExec().SetHost(hostIP)
	defer exec.Close()

	output, err := exec.Command("cat /proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}

	fields := strings.Fields(string(output))
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf("invalid loadavg: %s", string(output))
	}

	l1, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, 0, err
	}
	l5, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, 0, err
	}
	l15, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return l1, l5, l15, nil
}

func collectMemoryUsage(hostIP string) (uint64, uint64, float64, error) {
	exec := NewExec().SetHost(hostIP)
	defer exec.Close()

	output, err := exec.Command("cat /proc/meminfo")
	if err != nil {
		return 0, 0, 0, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	var memTotal, memAvailable uint64

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				v, err := strconv.ParseUint(fields[1], 10, 64)
				if err == nil {
					memTotal = v * 1024
				}
			}
		}
		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				v, err := strconv.ParseUint(fields[1], 10, 64)
				if err == nil {
					memAvailable = v * 1024
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, 0, err
	}

	if memTotal == 0 {
		return 0, 0, 0, fmt.Errorf("memtotal is zero")
	}

	memUsed := memTotal - memAvailable
	usage := float64(memUsed) / float64(memTotal) * 100
	return memTotal, memUsed, usage, nil
}

func collectRootDiskUsage(hostIP string) (uint64, uint64, float64, error) {
	exec := NewExec().SetHost(hostIP)
	defer exec.Close()

	output, err := exec.Command("df -B1 / | tail -n +2")
	if err != nil {
		return 0, 0, 0, err
	}

	line := util.Trim(string(output))
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return 0, 0, 0, fmt.Errorf("invalid df output: %s", line)
	}

	total, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	used, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	if total == 0 {
		return 0, 0, 0, nil
	}

	usage := float64(used) / float64(total) * 100
	return total, used, usage, nil
}

func collectDataDiskUsages(hostIP string) ([]DiskUsage, error) {
	disks, err := DescribeDisk(hostIP)
	if err != nil {
		return nil, err
	}
	points, err := DescribeMountedPoint(hostIP)
	if err != nil {
		return nil, err
	}

	diskByDevice := make(map[string]model.DiskDevice)
	for _, d := range disks {
		diskByDevice[d.Name] = d
	}

	result := make([]DiskUsage, 0)
	for _, p := range points {
		if !strings.HasPrefix(p.Point, "/mnt/") {
			continue
		}
		deviceDisk, ok := diskByDevice[p.Device]
		if !ok {
			continue
		}
		isHDD := deviceDisk.Rota
		usage, err := getDiskUsageWithCache(hostIP, p.Point, p.Device, isHDD)
		if err != nil {
			flog.Warnf("collect disk usage failed on host %s, point %s: %v", hostIP, p.Point, err)
			continue
		}

		usage.IsSystemDisk = deviceDisk.IsSystemDisk
		usage.IsHDD = isHDD
		usage.Filesystem = p.FsType
		usage.Device = p.Device
		usage.SpecMountPath = deviceDisk.SpecMountPoint

		result = append(result, usage)
	}

	return result, nil
}

func getDiskUsageWithCache(hostIP, mountPoint, device string, isHDD bool) (DiskUsage, error) {
	key := hostIP + "|" + mountPoint

	now := time.Now()
	hddInterval := 5 * time.Minute
	ssdInterval := 30 * time.Second
	interval := ssdInterval
	if isHDD {
		interval = hddInterval
	}

	diskUsageCache.mu.Lock()
	entry, ok := diskUsageCache.data[key]
	if ok && now.Sub(entry.lastSample) < interval {
		usage := entry.usage
		diskUsageCache.mu.Unlock()
		return usage, nil
	}
	diskUsageCache.mu.Unlock()

	exec := NewExec().SetHost(hostIP)
	defer exec.Close()

	cmd := fmt.Sprintf("df -B1 %s | tail -n +2", mountPoint)
	output, err := exec.Command(cmd)
	if err != nil {
		return DiskUsage{}, err
	}

	line := util.Trim(string(output))
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return DiskUsage{}, fmt.Errorf("invalid df output for %s: %s", mountPoint, line)
	}

	total, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return DiskUsage{}, err
	}
	used, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return DiskUsage{}, err
	}

	usagePercent := 0.0
	if total > 0 {
		usagePercent = float64(used) / float64(total) * 100
	}

	usage := DiskUsage{
		MountPoint:   mountPoint,
		TotalBytes:   total,
		UsedBytes:    used,
		UsagePercent: usagePercent,
	}

	diskUsageCache.mu.Lock()
	diskUsageCache.data[key] = diskUsageCacheEntry{
		usage:      usage,
		lastSample: now,
	}
	diskUsageCache.mu.Unlock()

	return usage, nil
}

func collectSambaMetrics(hostIP string) (ServiceMetrics, error) {
	cmd := NewExec().SetHost(hostIP)
	defer cmd.Close()

	checkInstalled := `
        if command -v smbd >/dev/null 2>&1; then
            echo "installed"
        else
            echo "not_installed"
        fi`

	resBs, err := cmd.Command(checkInstalled)
	if err != nil {
		return ServiceMetrics{}, err
	}

	installed := util.Trim(string(resBs)) == "installed"
	if !installed {
		return ServiceMetrics{
			Installed:         false,
			Status:            "not_installed",
			ActiveConnections: 0,
		}, nil
	}

	activeOutput, err := cmd.CommandWithoutExitCode("systemctl is-active smbd")
	if err != nil {
		return ServiceMetrics{}, err
	}
	activeState := util.Trim(string(activeOutput))

	status := "unknown"
	switch activeState {
	case "active":
		status = "running"
	case "inactive":
		status = "stopped"
	case "failed":
		status = "failed"
	default:
		status = "unknown"
	}

	connOutput, err := cmd.CommandWithoutExitCode("smbstatus -b 2>/dev/null | tail -n +5 | wc -l")
	if err != nil {
		flog.Warnf("collect samba connections failed on host %s: %v", hostIP, err)
		return ServiceMetrics{
			Installed:         installed,
			Status:            status,
			ActiveConnections: 0,
		}, nil
	}

	connStr := util.Trim(string(connOutput))
	conns, err := strconv.ParseInt(connStr, 10, 64)
	if err != nil {
		flog.Warnf("parse samba connection count failed on host %s: %v", hostIP, err)
		conns = 0
	}

	return ServiceMetrics{
		Installed:         installed,
		Status:            status,
		ActiveConnections: conns,
	}, nil
}

func collectNFSMetrics(hostIP string) (ServiceMetrics, error) {
	installed, _, serviceStatus, err := CheckNFSGaneshaInstallation(hostIP)
	if err != nil {
		return ServiceMetrics{}, err
	}

	status := serviceStatus
	if status == "" {
		status = "unknown"
	}

	cmd := NewExec().SetHost(hostIP)
	defer cmd.Close()

	connOutput, err := cmd.CommandWithoutExitCode("ss -tna 2>/dev/null | grep ':2049' | wc -l")
	if err != nil {
		flog.Warnf("collect nfs connections failed on host %s: %v", hostIP, err)
		return ServiceMetrics{
			Installed:         installed,
			Status:            status,
			ActiveConnections: 0,
		}, nil
	}

	connStr := util.Trim(string(connOutput))
	conns, err := strconv.ParseInt(connStr, 10, 64)
	if err != nil {
		flog.Warnf("parse nfs connection count failed on host %s: %v", hostIP, err)
		conns = 0
	}

	return ServiceMetrics{
		Installed:         installed,
		Status:            status,
		ActiveConnections: conns,
	}, nil
}
