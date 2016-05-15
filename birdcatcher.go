//  Copyright (c) 2015 Rackspace
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
//  implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package birdcatcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/openstack/swift/go/hummingbird"
)

func Lala() {
	fmt.Println("lalala")
}

type BirdCatcher struct {
	oring hummingbird.Ring
}

type DeviceData struct {
	Device  hummingbird.Device
	Mounted bool
}

type ReconData struct {
	Device  string
	Mounted bool
	dev     hummingbird.Device
}

func (bc *BirdCatcher) AllDevs() (devs []hummingbird.Device) {
	return bc.oring.AllDevices()
}

func (bc *BirdCatcher) GatherReconData() { //(devs []DeviceData, errs []error) {

	allWeightedDevs := make(map[string]*hummingbird.Device)
	allServers := make(map[string]bool)

	for _, dev := range bc.oring.AllDevices() {
		if dev.Weight > 0 {
			allWeightedDevs[fmt.Sprintf(
				"%s:%d/%s", dev.Ip, dev.Port, dev.Device)] = &dev
			serverUrl := fmt.Sprintf("%s:%d", dev.Ip, dev.Port)
			if _, ok := allServers[serverUrl]; !ok {
				allServers[serverUrl] = true
			}
		}
	}

	//var allReconData []ReconData

	for hostname, _ := range allServers {
		fmt.Println("111")
		serverUrl := fmt.Sprintf("http://%s/recon/unmounted", hostname)

		fmt.Println(serverUrl)
		resp, err := http.Get(serverUrl)
		if err != nil {
			//errs := append(errs, err)
			//serverToDev[serverUrl] = nil
			fmt.Println("333: ", err)
			continue
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			//errs := append(errs, err)
			fmt.Println("444", err)
			continue
		}
		var allReconData []*ReconData
		if err := json.Unmarshal(data, &allReconData); err != nil {
			//errs := append(errs, err)
			fmt.Println("555", err)
			continue
		}

		fmt.Println("aaaaa: ", allReconData[0])

		for _, rData := range allReconData {
			devKey := fmt.Sprintf("%s/%s", hostname, rData.Device)
			if wDev, ok := allWeightedDevs[devKey]; ok {
				rData.dev = *wDev
			} else {
				fmt.Println("could not find dev in dict")
			}
		}
		fmt.Println("777777: ", allReconData[0].Device)
		fmt.Println("777777: ", allReconData[0].Mounted)
		fmt.Println("777777: ", allReconData[0].dev.Device)
	}
}

func GetBirdCatcher() (*BirdCatcher, error) {

	hashPathPrefix, hashPathSuffix, err := hummingbird.GetHashPrefixAndSuffix()
	if err != nil {
		fmt.Println("Unable to load hash path prefix and suffix:", err)
		return nil, err
	}
	objRing, err := hummingbird.GetRing("object", hashPathPrefix, hashPathSuffix)
	if err != nil {
		fmt.Println("Unable to load ring:", err)
		return nil, err
	}
	bc := BirdCatcher{}
	bc.oring = objRing
	return &bc, nil

}
