package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// This would take a snapshot of the VM state, stop the vm and save location of snap
func stopMachine(c echo.Context) error {
	machineID := c.Param("machine_id")
	vm, _ := fcManager.GetVM(machineID)

	fmt.Println("Stopping VM ...")
	// if err := vm.machine.Shutdown(vm.vmmCtx); err != nil {
	// 	return c.JSON(http.StatusInternalServerError, fmt.Sprintf("failed to stop machine: %v", err))
	// }

	ctx := context.Background()

	vmmCtx, vmmCancel := context.WithCancel(ctx)

	if err := vm.machine.Shutdown(vmmCtx); err != nil {
		vmmCancel()
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("failed to stop machine: %v", err))
	}

	updateMachineStatus(vm.vmmCtx, machineID, StatusStopped)

	return c.JSON(http.StatusOK, "Machine stopped!")
}

// This would use the snapshot to start the VM
func startMachine(c echo.Context) error {
	machineID := c.Param("machine_id")
	vm, _ := fcManager.GetVM(machineID)

	fmt.Println("Starting VM ...")
	if err := vm.machine.Start(vm.vmmCtx); err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("failed to start machine: %v", err))
	}

	updateMachineStatus(vm.vmmCtx, machineID, StatusRunning)

	return c.JSON(http.StatusOK, "Machine restarted!")
}

// This would delete the vm deets
func deleteMachine(c echo.Context) error {
	// machineID := c.Param("machine_id")
	// vm, _ := fcManager.GetVM(machineID)

	// memFilePath := "path/to/memory/file"
	// snapshotPath := "path/to/snapshot/file"

	// fmt.Println("Restarting VM from snapshot...")

	// err := vm.machine.StopVMM(vm.vmmCtx, memFilePath, snapshotPath)
	// if err != nil {
	// 	return fmt.Errorf("failed to create snapshot: %v", err)
	// }

	// updateMachineStatus(vm.vmmCtx, machineID, StatusStopped)

	return c.JSON(http.StatusOK, "Machine deleted!")
}
