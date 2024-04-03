package main

import "sync"

type FirecrackerManager struct {
	sync.Mutex
	vms map[string]*runningFirecracker
}

func NewFirecrackerManager() *FirecrackerManager {
	return &FirecrackerManager{
		vms: make(map[string]*runningFirecracker),
	}
}

func (manager *FirecrackerManager) AddVM(id string, vm *runningFirecracker) {
	manager.Lock()
	defer manager.Unlock()
	manager.vms[id] = vm
}

func (manager *FirecrackerManager) RemoveVM(id string) {
	manager.Lock()
	defer manager.Unlock()
	delete(manager.vms, id)
}

func (manager *FirecrackerManager) GetVM(id string) (*runningFirecracker, bool) {
	manager.Lock()
	defer manager.Unlock()
	vm, exists := manager.vms[id]
	return vm, exists
}
