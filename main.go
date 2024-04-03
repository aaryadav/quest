package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func defaultMachineConfig() *ApiMachineConfig {
	return &ApiMachineConfig{
		AppName: "crunchy_new_app",
		Image:   "default_image",
		MachineType: ApiMachineType{
			CpuKind:  "default_cpu",
			Cpus:     1,
			GpuKind:  "default_gpu",
			MemoryMb: 256,
		},
	}
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	e := echo.New()

	// Define the routes
	// e.POST("/apps/:app_name/machines", createMachine)
	// e.GET("/apps/:app_name/machines/:machine_id/wait", waitForMachineState)
	// e.GET("/apps/:app_name/machines/:machine_id", getMachine)
	// e.GET("/apps/machines", listMachines)
	// e.POST("/apps/:app_name/machines/:machine_id/run", runCode)

	e.POST("/machines", createMachine)
	e.GET("/machines/:machine_id/wait", waitForMachineState)
	e.GET("/machines/:machine_id", getMachine)
	e.GET("/machines", listMachines)
	e.POST("/machines/:machine_id/run", runCode)

	e.GET("/machines/:machine_id/start", startMachine)
	e.GET("/machines/:machine_id/stop", stopMachine)
	e.GET("/machines/:machine_id/delete", startMachine)

	// Start the server
	e.Logger.Fatal(e.Start(":1323"))
}

var fcManager = NewFirecrackerManager()
var client = &http.Client{}

func handleError(c echo.Context, err error, httpStatus int, errMsg string) error {
	log.WithError(err).Error(errMsg)
	return c.JSON(httpStatus, map[string]string{"error": errMsg})
}

func fetchMachineInfo(ctx context.Context, machineID string) (*MachineInfo, error) {
	data, err := rdb.Get(ctx, "machine:"+machineID).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("machine not found")
	} else if err != nil {
		return nil, err
	}

	var machineInfo MachineInfo
	if err = json.Unmarshal([]byte(data), &machineInfo); err != nil {
		return nil, err
	}

	return &machineInfo, nil
}

func createAndInitializeVM(ctx context.Context, machineConfig *ApiMachineConfig) (*runningFirecracker, error) {
	// machineConfig := defaultMachineConfig()
	vm, err := createAndStartVM(ctx, machineConfig)
	if err != nil {
		log.WithError(err).Error("failed to create VMM")
		updateMachineStatus(ctx, vm.vmmID, StatusFailed)
		return nil, err
	}

	log.WithField("ip", vm.ip).Info("New VM created and started")
	fcManager.AddVM(vm.vmmID, vm)
	go healthCheckMachine(ctx, vm.ip, vm.vmmID)

	return vm, nil
}

func createMachine(c echo.Context) error {
	ctx := context.Background()
	machineConfig := defaultMachineConfig()

	log.Info(machineConfig)

	vm, err := createAndInitializeVM(ctx, machineConfig)
	if err != nil {
		return handleError(c, err, http.StatusInternalServerError, "Failed to create and initialize VM")
	}

	return c.JSON(http.StatusOK, CreateMachineResponse{
		MachineID:     vm.vmmID,
		IP:            vm.ip,
		MachineConfig: *defaultMachineConfig(),
	})
}

func waitForMachineState(c echo.Context) error {
	machineID := c.Param("machine_id")
	ctx := context.Background()

	machineInfo, err := fetchMachineInfo(ctx, machineID)
	if err != nil {
		if strings.Contains(err.Error(), "machine not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, MachineStatusResponse{
		MachineID: machineID,
		Status:    MachineStatusType(machineInfo.Status),
	})
}

func getMachine(c echo.Context) error {
	machineID := c.Param("machine_id")
	ctx := context.Background()

	machineInfo, err := fetchMachineInfo(ctx, machineID)
	if err != nil {
		if strings.Contains(err.Error(), "machine not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	machineConfig := defaultMachineConfig()

	return c.JSON(http.StatusOK, CreateMachineResponse{
		MachineID:     machineID,
		MachineConfig: *machineConfig,
		Status:        MachineStatusType(machineInfo.Status),
	})
}

func listMachines(c echo.Context) error {
	ctx := context.Background()

	// Assuming all machine IDs are stored with a common prefix, e.g., "machine:"
	keys, err := rdb.Keys(ctx, "machine:*").Result()
	if err != nil {
		log.WithError(err).Error("failed to fetch machine keys from Redis")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	var machines []CreateMachineResponse

	for _, key := range keys {
		data, err := rdb.Get(ctx, key).Result()
		if err != nil {
			log.WithError(err).Errorf("failed to fetch machine info for key %s", key)
			continue // Skip this machine and continue with others
		}

		var machineInfo MachineInfo
		err = json.Unmarshal([]byte(data), &machineInfo)
		if err != nil {
			log.WithError(err).Errorf("failed to unmarshal machine info for key %s", key)
			continue // Skip this machine and continue with others
		}

		machineID := strings.TrimPrefix(key, "machine:") // Remove prefix to get the actual ID
		machineConfig := defaultMachineConfig()          // Or fetch specific config if stored separately

		machines = append(machines, CreateMachineResponse{
			MachineID:     machineID,
			MachineConfig: *machineConfig,
			Status:        MachineStatusType(machineInfo.Status),
			// IP will be empty if not available
			IP: net.ParseIP(machineInfo.IP),
		})
	}

	return c.JSON(http.StatusOK, machines)
}

func runCode(c echo.Context) error {
	machineID := c.Param("machine_id")
	ctx := context.Background()

	machineInfo, err := fetchMachineInfo(ctx, machineID)
	if err != nil {
		if strings.Contains(err.Error(), "machine not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	machineIP := machineInfo.IP
	var codeRunRequest CodeRunRequest

	if err := c.Bind(&codeRunRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	url := fmt.Sprintf("http://%s:8081/run", machineIP)
	jsonData, err := json.Marshal(codeRunRequest)
	if err != nil {
		log.WithError(err).Error("failed to marshal code run request")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to send request to machine")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var codeRunResponse CodeRunResponse
	err = json.Unmarshal(body, &codeRunResponse)
	if err != nil {
		log.WithError(err).Error("failed to unmarshal code run response")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, codeRunResponse)
}
