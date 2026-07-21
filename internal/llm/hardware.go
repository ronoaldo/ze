package llm

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// ModelConfig represents the configuration parameters for a specific model size.
type ModelConfig struct {
	Name        string
	MinVRAMMB   int64
	RequiresGPU bool
}

var (
	// Known Gemma 4 models and their approximate VRAM requirements in MB
	Gemma4Models = []ModelConfig{
		{Name: "gemma-4-27b", MinVRAMMB: 18000, RequiresGPU: true}, // High end
		{Name: "gemma-4-9b", MinVRAMMB: 6000, RequiresGPU: false},  // Mid range (can run on CPU)
		{Name: "gemma-4-2b", MinVRAMMB: 2000, RequiresGPU: false},  // Low end
	}
)

// HardwareResources describes the detected hardware capabilities.
type HardwareResources struct {
	HasGPU      bool
	GPUMemoryMB int64
	CPUCores    int
}

// DetectHardware attempts to detect GPU and CPU information.
func DetectHardware() *HardwareResources {
	res := &HardwareResources{
		HasGPU:   false,
		CPUCores: runtime.NumCPU(),
	}

	if isLinux() || isDarwin() {
		detectGPUMemory(res)
	}

	return res
}

func isLinux() bool  { return runtime.GOOS == "linux" }
func isDarwin() bool { return runtime.GOOS == "darwin" }

// detectGPUMemory tries to find GPU memory using nvidia-smi if available (common on Linux).
func detectGPUMemory(res *HardwareResources) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=memory.total", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err == nil {
		valStr := strings.TrimSpace(string(output))
		lines := strings.Split(valStr, "\n")
		if len(lines) > 0 {
			val, err := strconv.ParseInt(strings.TrimSpace(lines[0]), 10, 64)
			if err == nil {
				res.GPUMemoryMB = val
				res.HasGPU = true
			}
		}
	}
}

// SelectBestModel chooses the best Gemma 4 model based on detected hardware resources.
func SelectBestModel(res *HardwareResources) string {
	var selected string

	for _, m := range Gemma4Models {
		if res.HasGPU && res.GPUMemoryMB >= m.MinVRAMMB {
			return m.Name // Return the first one that fits (assuming list is ordered largest to smallest)
		}
		// If we don't have a GPU or not enough VRAM, check if it can run on CPU
		if !m.RequiresGPU && res.CPUCores >= 4 {
			selected = m.Name // Keep track of the best non-GPU model found so far
		} else if !m.RequiresGPU && selected == "" {
			// Fallback to smallest if nothing else fits
			selected = m.Name
		}
	}

	if selected == "" {
		return Gemma4Models[len(Gemma4Models)-1].Name // Default to smallest
	}

	return selected
}
