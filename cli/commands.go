package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Creates/Starts a microVM",
	Args:  cobra.ExactArgs(1),
	Run:   startMachine,
}
var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stops a microVM",
	Args:  cobra.ExactArgs(1),
	Run:   stopMachine,
}
var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Status and Specs of the microVM",
	Args:  cobra.ExactArgs(1),
	Run:   getMachine,
}
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all microVMs",
	Run:   listMachines,
}

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Deletes a microvm",
	Args:  cobra.ExactArgs(1),
	Run:   deleteMachine,
}

func startMachine(cmd *cobra.Command, args []string) {
	machineID := args[0]
	fmt.Printf("Starting machine '%s'...\n", machineID)

	resp, err := makeRequest("GET", fmt.Sprintf("/machines/%s/start", machineID), nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var startMachineResponse StartMachineResponse
	if err := json.NewDecoder(resp.Body).Decode(&startMachineResponse); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return
	}

	prettyPrintOutput(startMachineResponse)
}

func getMachine(cmd *cobra.Command, args []string) {
	machineID := args[0]
	fmt.Printf("Status of '%s':\n", machineID)

	resp, err := makeRequest("GET", fmt.Sprintf("/machines/%s", machineID), nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	defer resp.Body.Close()

	var createMachineResponse CreateMachineResponse
	if err := json.NewDecoder(resp.Body).Decode(&createMachineResponse); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return
	}

	prettyPrintOutput(createMachineResponse)
}

func deleteMachine(cmd *cobra.Command, args []string) {
	machineID := args[0]
	fmt.Printf("Stopping and deleting '%s'\n", machineID)

	resp, err := makeRequest("DELETE", fmt.Sprintf("/machines/%s", machineID), nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var deleteMachineResponse DeleteMachineResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteMachineResponse); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return
	}

	prettyPrintOutput(deleteMachineResponse)
}

func stopMachine(cmd *cobra.Command, args []string) {
	machineID := args[0]
	fmt.Printf("Stopping machine '%s'...\n", machineID)

	resp, err := makeRequest("GET", fmt.Sprintf("/machines/%s/stop", machineID), nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var stopMachineResponse StopMachineResponse
	if err := json.NewDecoder(resp.Body).Decode(&stopMachineResponse); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return
	}

	prettyPrintOutput(stopMachineResponse)
}

func listMachines(cmd *cobra.Command, args []string) {

	resp, err := makeRequest("GET", "/machines", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var machineList []CreateMachineResponse
	if err := json.NewDecoder(resp.Body).Decode(&machineList); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return
	}

	for _, value := range machineList {
		prettyPrintOutput(value)
	}

}
