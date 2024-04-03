package main

import (
	"net"
)

type ApiMachineConfig struct {
	AppName     string         `json:"app_name"`
	Image       string         `json:"image"`
	MachineType ApiMachineType `json:"machine_type"`
}

type ApiMachineType struct {
	CpuKind  string `json:"cpu_kind"`
	Cpus     int32  `json:"cpus"`
	GpuKind  string `json:"gpu_kind"`
	MemoryMb int32  `json:"memory_mb"`
}

type CreateMachineResponse struct {
	MachineID     string            `json:"machine_id"`
	IP            net.IP            `json:"ip,omitempty"`
	Status        MachineStatusType `json:"status,omitempty"`
	MachineConfig ApiMachineConfig  `json:"machine_config"`
}

type MachineStatusResponse struct {
	MachineID string            `json:"machine_id"`
	Status    MachineStatusType `json:"status"`
}

type MachineStatusType string

const (
	StatusPending   MachineStatusType = "pending"
	StatusRunning   MachineStatusType = "running"
	StatusStopped   MachineStatusType = "stopped"
	StatusFailed    MachineStatusType = "failed"
	StatusCompleted MachineStatusType = "completed"
)

type CodeRunRequest struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Language string `json:"language"`
	Variant  string `json:"variant"`
}

type CodeRunResponse struct {
	Message      string `json:"message"`
	Error        string `json:"error"`
	Stdout       string `json:"stdout"`
	Stderr       string `json:"stderr"`
	ExecDuration int    `json:"exec_duration"`
	MemUsage     int    `json:"mem_usage"`
}

type DeleteMachineResponse struct{}

type StopMachineResponse struct{}

type StartMachineResponse struct{}
