/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gceGCEDriver

import (
	"sync"
	"time"

	"k8s.io/klog/v2"
)

// deviceErrMap is an atomic data datastructure for recording deviceInUseError times
// for specified devices
type deviceErrMap struct {
	enabled           bool
	timeout           time.Duration
	mux               sync.Mutex
	deviceInUseErrors map[string]time.Time
}

func newDeviceErrMap(shouldEnable bool, timeout time.Duration) *deviceErrMap {
	return &deviceErrMap{
		deviceInUseErrors: make(map[string]time.Time),
		enabled:           shouldEnable,
		timeout:           timeout,
	}
}

// checkDeviceErrorTimeout returns true an error was encountered for the specified deviceName,
// where the error happened at least `deviceInUseTimeout` seconds ago.
func (devErrMap *deviceErrMap) checkDeviceErrorTimeout(deviceName string) bool {
	if !devErrMap.enabled {
		return false
	}

	devErrMap.mux.Lock()
	defer devErrMap.mux.Unlock()

	lastErrTime, exists := devErrMap.deviceInUseErrors[deviceName]
	return exists && time.Now().Sub(lastErrTime).Seconds() >= devErrMap.timeout.Seconds()
}

// markDeviceError updates the internal `deviceInUseErrors` map to denote an error was encounted
// for the specified deviceName at the current time. If an error had previously been recorded, the
// time will not be updated.
func (devErrMap *deviceErrMap) markDeviceError(deviceName string) {
	if !devErrMap.enabled {
		return
	}

	devErrMap.mux.Lock()
	defer devErrMap.mux.Unlock()

	// If an earlier error has already been recorded, do not overwrite it
	if _, exists := devErrMap.deviceInUseErrors[deviceName]; !exists {
		now := time.Now()
		klog.V(4).Infof("Recording in-use error for device %s at time %s", deviceName, now)
		devErrMap.deviceInUseErrors[deviceName] = now
	}
}

// deleteDevice removes a specified device name from the map
func (devErrMap *deviceErrMap) deleteDevice(deviceName string) {
	if !devErrMap.enabled {
		return
	}

	devErrMap.mux.Lock()
	defer devErrMap.mux.Unlock()
	delete(devErrMap.deviceInUseErrors, deviceName)
}
