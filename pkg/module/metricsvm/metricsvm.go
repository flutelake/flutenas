package metricsvm

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"flutelake/fluteNAS/pkg/module/flog"

	"github.com/VictoriaMetrics/metrics"
)

type nodeValues struct {
	CPUUsagePercent  float64
	MemUsagePercent  float64
	RootUsagePercent float64
	Load1            float64
	Load5            float64
	Load15           float64
}

type diskValues struct {
	hostIP       string
	mountPoint   string
	device       string
	filesystem   string
	systemDisk   bool
	hdd          bool
	UsagePercent float64
	TotalBytes   uint64
	UsedBytes    uint64
}

type serviceValues struct {
	hostIP            string
	service           string
	installed         bool
	status            string
	ActiveConnections int64
}

var (
	initOnce     sync.Once
	nodeMu       sync.RWMutex
	nodeByHost   = make(map[string]nodeValues)
	diskMu       sync.RWMutex
	diskByKey    = make(map[string]diskValues)
	serviceMu    sync.RWMutex
	serviceByKey = make(map[string]serviceValues)
)

func Init() {
	initOnce.Do(func() {
		metrics.RegisterMetricsWriter(writeMetrics)
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	Init()
	metrics.WritePrometheus(w, true)
}

func InitPushFromEnv() {
	pushURL := os.Getenv("VICTORIA_METRICS_PUSH_URL")
	if pushURL == "" {
		pushURL = "http://127.0.0.1:8086/api/v1/import/prometheus"
	}
	hostname, err := os.Hostname()
	if err != nil {
		flog.Warnf("get hostname failed for victoria metrics push: %v", err)
		hostname = "unknown"
	}
	extraLabels := os.Getenv("VICTORIA_METRICS_EXTRA_LABELS")
	if extraLabels == "" {
		extraLabels = fmt.Sprintf(`instance="%s"`, hostname)
	}
	err = metrics.InitPush(pushURL, 10*time.Second, extraLabels, true)
	if err != nil {
		flog.Warnf("init victoria metrics push failed: %v", err)
	}
}

func UpdateNodeMetrics(hostIP string, cpuUsagePercent float64, memUsagePercent float64, rootUsagePercent float64, load1 float64, load5 float64, load15 float64) {
	nodeMu.Lock()
	nodeByHost[hostIP] = nodeValues{
		CPUUsagePercent:  cpuUsagePercent,
		MemUsagePercent:  memUsagePercent,
		RootUsagePercent: rootUsagePercent,
		Load1:            load1,
		Load5:            load5,
		Load15:           load15,
	}
	nodeMu.Unlock()
}

func UpdateDiskUsage(hostIP string, mountPoint string, device string, filesystem string, isSystemDisk bool, isHDD bool, usagePercent float64, totalBytes uint64, usedBytes uint64) {
	key := fmt.Sprintf("%s|%s|%s|%s", hostIP, mountPoint, device, filesystem)
	diskMu.Lock()
	diskByKey[key] = diskValues{
		hostIP:       hostIP,
		mountPoint:   mountPoint,
		device:       device,
		filesystem:   filesystem,
		systemDisk:   isSystemDisk,
		hdd:          isHDD,
		UsagePercent: usagePercent,
		TotalBytes:   totalBytes,
		UsedBytes:    usedBytes,
	}
	diskMu.Unlock()
}

func UpdateServiceMetrics(hostIP string, service string, installed bool, status string, activeConnections int64) {
	key := fmt.Sprintf("%s|%s", hostIP, service)
	serviceMu.Lock()
	serviceByKey[key] = serviceValues{
		hostIP:            hostIP,
		service:           service,
		installed:         installed,
		status:            status,
		ActiveConnections: activeConnections,
	}
	serviceMu.Unlock()
}

func writeMetrics(w io.Writer) {
	nodeMu.RLock()
	for host, v := range nodeByHost {
		n, err := fmt.Fprintf(w, "flutenas_node_cpu_usage_percent{host=%q} %g\n", host, v.CPUUsagePercent)
		if err != nil {
			flog.Errorf("write metrics %d failed: %v", n, err)
			continue
		}
		fmt.Fprintf(w, "flutenas_node_mem_usage_percent{host=%q} %g\n", host, v.MemUsagePercent)
		fmt.Fprintf(w, "flutenas_node_root_usage_percent{host=%q} %g\n", host, v.RootUsagePercent)
		fmt.Fprintf(w, "flutenas_node_load1{host=%q} %g\n", host, v.Load1)
		fmt.Fprintf(w, "flutenas_node_load5{host=%q} %g\n", host, v.Load5)
		fmt.Fprintf(w, "flutenas_node_load15{host=%q} %g\n", host, v.Load15)
	}
	nodeMu.RUnlock()

	diskMu.RLock()
	for _, v := range diskByKey {
		systemValue := "false"
		if v.systemDisk {
			systemValue = "true"
		}
		hddValue := "false"
		if v.hdd {
			hddValue = "true"
		}
		fmt.Fprintf(
			w,
			"flutenas_data_disk_usage_percent{host=%q,mount_point=%q,device=%q,filesystem=%q,system_disk=%q,hdd=%q} %g\n",
			v.hostIP,
			v.mountPoint,
			v.device,
			v.filesystem,
			systemValue,
			hddValue,
			v.UsagePercent,
		)
		fmt.Fprintf(
			w,
			"flutenas_data_disk_total_bytes{host=%q,mount_point=%q,device=%q,filesystem=%q,system_disk=%q,hdd=%q} %d\n",
			v.hostIP,
			v.mountPoint,
			v.device,
			v.filesystem,
			systemValue,
			hddValue,
			v.TotalBytes,
		)
		fmt.Fprintf(
			w,
			"flutenas_data_disk_used_bytes{host=%q,mount_point=%q,device=%q,filesystem=%q,system_disk=%q,hdd=%q} %d\n",
			v.hostIP,
			v.mountPoint,
			v.device,
			v.filesystem,
			systemValue,
			hddValue,
			v.UsedBytes,
		)
	}
	diskMu.RUnlock()

	serviceMu.RLock()
	for _, v := range serviceByKey {
		installedValue := "false"
		if v.installed {
			installedValue = "true"
		}
		fmt.Fprintf(
			w,
			"flutenas_service_active_connections{host=%q,service=%q,status=%q,installed=%q} %d\n",
			v.hostIP,
			v.service,
			v.status,
			installedValue,
			v.ActiveConnections,
		)
	}
	serviceMu.RUnlock()
}
