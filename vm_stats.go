package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type MachineInfo struct {
	Status string `json:"status"`
	IP     string `json:"ip,omitempty"`
}

func updateMachineStatus(ctx context.Context, machineID string, newStatus MachineStatusType, ip ...net.IP) {
	info := MachineInfo{
		Status: string(newStatus),
	}

	if len(ip) > 0 {
		info.IP = ip[0].String()
	}

	data, err := json.Marshal(info)
	if err != nil {
		log.WithError(err).Error("failed to marshal machine info")
		return
	}

	err = rdb.Set(ctx, "machine:"+machineID, data, 0).Err()

	if err != nil {
		log.WithError(err).Error("failed to update machine status in Redis")
	} else {
		log.Infof("Updated status of machine %s in Redis", machineID)
	}
}

func healthCheckMachine(ctx context.Context, machineIP net.IP, machineID string) {
	url := "http://" + machineIP.String() + ":8081/health"

	for i := 0; i < HealthCheckMaxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			log.Errorf("Health check failed for machine %s: %v", machineID, err)
			time.Sleep(HealthCheckInterval)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Infof("Machine %s is healthy", machineID)
			updateMachineStatus(ctx, machineID, StatusRunning, machineIP)
			return
		}

		log.Warnf("Machine %s is not ready, retrying...", machineID)
		time.Sleep(HealthCheckInterval)
	}

	log.Errorf("Machine %s failed to become healthy after retries", machineID)
	updateMachineStatus(ctx, machineID, StatusFailed)
}
