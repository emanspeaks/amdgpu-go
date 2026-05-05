//go:build ignore

package amdgpu_test

import (
	"fmt"
	"log"

	"github.com/emanspeaks/amdgpu-go/amdgpu"
)

// ExampleOpen shows how to open the first AMD GPU, query its device info,
// print the marketing name, and close the device.
func ExampleOpen() {
	dev, err := amdgpu.Open(0)
	if err != nil {
		log.Fatalf("open GPU 0: %v", err)
	}
	defer dev.Close()

	info, err := dev.DeviceInfo()
	if err != nil {
		log.Fatalf("DeviceInfo: %v", err)
	}
	fmt.Println(info.MarketingName)
}
