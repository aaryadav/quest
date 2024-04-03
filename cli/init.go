package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/spf13/cobra"
)

type CreateMachineOutput struct {
	MachineID     string           `json:"machine_id"`
	IP            net.IP           `json:"ip,omitempty"`
	MachineConfig ApiMachineConfig `json:"machine_config"`
}

var initCmd = &cobra.Command{
	Use:   "init [name] [config]",
	Short: "Initialize and start a new microVM",
	// Args:  cobra.ExactArgs(2),
	Args: cobra.MinimumNArgs(1),
	Run:  createMachine,
}

func createMachine(cmd *cobra.Command, args []string) {
	machineID := args[0]
	if len(args) > 1 {
		configFile := args[1]
		fmt.Printf("Starting machine '%s' with config file '%s'\n", machineID, configFile)
	}
	fmt.Printf("Starting machine '%s' with no config file\n", machineID)

	resp, err := makeRequest("POST", "/machines", nil)
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

func prettyPrintOutput(v interface{}) {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling output:", err)
		return
	}

	fmt.Println(string(jsonBytes))
}

func makeRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, "http://localhost:1323"+path, body)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, fmt.Errorf("error making request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		var errResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("error unmarshaling error response: %v", err)
		}
		// Include the error message from the response in the returned error
		return nil, fmt.Errorf("error: %s", errResp["error"])
	}
	return resp, nil
}
