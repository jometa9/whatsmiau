package services

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/verbeux-ai/whatsmiau/env"
	"github.com/verbeux-ai/whatsmiau/interfaces"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type SystemMetrics struct {
	PublicIP        string              `json:"publicIp"`
	CPUUsage        float64             `json:"cpuUsage"`
	MemoryUsage     float64             `json:"memoryUsage"`
	Port            string              `json:"port"`
	ProcessCPU      float64             `json:"processCpu"`
	ProcessMemory   float64             `json:"processMemory"`
	Instances       []InstanceMetrics   `json:"instances"`
	InstancesCount  int                 `json:"instancesCount"`
}

type InstanceMetrics struct {
	InstanceID   string  `json:"instanceId"`
	CPUUsage     float64 `json:"cpuUsage"`
	MemoryUsage  float64 `json:"memoryUsage"`
}

var (
	monitorRunning      bool
	monitorMutex        sync.Mutex
	instanceRepo        interfaces.InstanceRepository
	connectedChecker    func(instanceID string) bool
)

// SetConnectedInstanceChecker sets a function to check if an instance is connected
// This avoids import cycle between services and whatsmiau packages
func SetConnectedInstanceChecker(checker func(instanceID string) bool) {
	monitorMutex.Lock()
	defer monitorMutex.Unlock()
	connectedChecker = checker
}

// SetInstanceRepository sets the instance repository for monitoring
func SetInstanceRepository(repo interfaces.InstanceRepository) {
	instanceRepo = repo
}

func StartMonitor() {
	monitorMutex.Lock()
	defer monitorMutex.Unlock()

	if monitorRunning {
		return
	}
	monitorRunning = true

	// Send immediately on startup
	go sendMetrics()

	// Send every minute
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			sendMetrics()
		}
	}()
}

func sendMetrics() {
	webhookURL := env.GetWebhookURL()
	if webhookURL == "" {
		return
	}

	metrics, err := getSystemMetrics()
	if err != nil {
		zap.L().Error("failed to get system metrics", zap.Error(err))
		return
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		zap.L().Error("failed to marshal metrics", zap.Error(err))
		return
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		zap.L().Error("failed to create request", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("failed to send metrics to webhook", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if env.Env.DebugMode {
			zap.L().Info("metrics sent successfully", zap.Any("metrics", metrics))
		}
	} else {
		body, _ := io.ReadAll(resp.Body)
		if env.Env.DebugMode {
			zap.L().Warn("webhook returned non-2xx status", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		}
	}
}

func getSystemMetrics() (*SystemMetrics, error) {
	// Get public IP
	publicIP, err := getPublicIP()
	if err != nil {
		zap.L().Warn("failed to get public IP", zap.Error(err))
		publicIP = "unknown"
	}

	// Get system CPU usage
	cpuUsage := getCPUUsage()

	// Get system memory usage
	memUsage := getMemoryUsage()

	// Get process metrics (Go process)
	processCPU, processMem := getProcessMetrics()

	// Get instances metrics
	instancesMetrics, instancesCount := getInstancesMetrics(processCPU, processMem)

	return &SystemMetrics{
		PublicIP:       publicIP,
		CPUUsage:       cpuUsage,
		MemoryUsage:    memUsage,
		Port:            env.Env.Port,
		ProcessCPU:     processCPU,
		ProcessMemory:  processMem,
		Instances:       instancesMetrics,
		InstancesCount: instancesCount,
	}, nil
}

func getPublicIP() (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func getCPUUsage() float64 {
	// Get real CPU usage from the system using gopsutil
	// This gets the average CPU usage across all cores
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		zap.L().Warn("failed to get CPU usage", zap.Error(err))
		return 0
	}

	if len(percentages) > 0 {
		cpuPercent := percentages[0]
		// Clamp to valid range
		if cpuPercent > 100 {
			cpuPercent = 100
		}
		if cpuPercent < 0 {
			cpuPercent = 0
		}
		return cpuPercent
	}

	return 0
}

func getMemoryUsage() float64 {
	// Get real memory usage from the system using gopsutil
	// This gets the total system memory usage percentage
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		zap.L().Warn("failed to get memory usage", zap.Error(err))
		return 0
	}

	// Return the percentage of used memory
	usage := vmStat.UsedPercent
	
	// Clamp to valid range
	if usage > 100 {
		usage = 100
	}
	if usage < 0 {
		usage = 0
	}
	
	return usage
}

func getProcessMetrics() (float64, float64) {
	// Get current process
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		zap.L().Warn("failed to get process", zap.Error(err))
		return 0, 0
	}

	// Get CPU usage percentage (averaged over 1 second)
	cpuPercent, err := proc.CPUPercent()
	if err != nil {
		zap.L().Warn("failed to get process CPU", zap.Error(err))
		cpuPercent = 0
	}

	// Get memory usage in MB
	memInfo, err := proc.MemoryInfo()
	if err != nil {
		zap.L().Warn("failed to get process memory", zap.Error(err))
		return cpuPercent, 0
	}

	// Get total system memory to calculate percentage
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		zap.L().Warn("failed to get system memory", zap.Error(err))
		return cpuPercent, 0
	}

	// Calculate memory usage percentage
	memPercent := (float64(memInfo.RSS) / float64(vmStat.Total)) * 100

	// Clamp values
	if cpuPercent > 100 {
		cpuPercent = 100
	}
	if cpuPercent < 0 {
		cpuPercent = 0
	}
	if memPercent > 100 {
		memPercent = 100
	}
	if memPercent < 0 {
		memPercent = 0
	}

	return cpuPercent, memPercent
}

func getInstancesMetrics(processCPU, processMem float64) ([]InstanceMetrics, int) {
	if instanceRepo == nil {
		// No repository set, return process metrics
		return []InstanceMetrics{
			{
				InstanceID:   "process",
				CPUUsage:     processCPU,
				MemoryUsage:  processMem,
			},
		}, 0
	}

	// Get all instances from repository
	ctx := context.Background()
	instancesList, err := instanceRepo.List(ctx, "")
	if err != nil {
		if env.Env.DebugMode {
			zap.L().Warn("failed to list instances for metrics", zap.Error(err))
		}
		return []InstanceMetrics{
			{
				InstanceID:   "process",
				CPUUsage:     processCPU,
				MemoryUsage:  processMem,
			},
		}, 0
	}

	instancesCount := len(instancesList)

	if instancesCount == 0 {
		// No instances, return process metrics
		return []InstanceMetrics{
			{
				InstanceID:   "process",
				CPUUsage:     processCPU,
				MemoryUsage:  processMem,
			},
		}, 0
	}

	// Check which instances are actually connected using the checker function
	// This avoids import cycle between services and whatsmiau packages
	connectedInstances := make(map[string]bool)
	connectedCount := 0

	if connectedChecker != nil {
		for _, inst := range instancesList {
			if connectedChecker(inst.ID) {
				connectedInstances[inst.ID] = true
				connectedCount++
			}
		}
	}

	// Calculate per-instance metrics
	// Only distribute process metrics among connected instances
	// If no instances are connected, all get 0
	var cpuPerInstance, memPerInstance float64
	if connectedCount > 0 {
		cpuPerInstance = processCPU / float64(connectedCount)
		memPerInstance = processMem / float64(connectedCount)
	} else {
		cpuPerInstance = 0
		memPerInstance = 0
	}

	// Build instances metrics
	instancesMetrics := make([]InstanceMetrics, 0, instancesCount)
	for _, inst := range instancesList {
		if connectedInstances[inst.ID] {
			// Instance is connected, assign its share of process resources
			instancesMetrics = append(instancesMetrics, InstanceMetrics{
				InstanceID:   inst.ID,
				CPUUsage:     cpuPerInstance,
				MemoryUsage:  memPerInstance,
			})
		} else {
			// Instance is not connected, assign 0
			instancesMetrics = append(instancesMetrics, InstanceMetrics{
				InstanceID:   inst.ID,
				CPUUsage:     0,
				MemoryUsage:  0,
			})
		}
	}

	return instancesMetrics, instancesCount
}

