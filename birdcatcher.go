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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/openstack/swift/go/hummingbird"
)

type BirdCatcher struct {
	oring  hummingbird.Ring
	logger hummingbird.SysLogLike
}

type ReconData struct {
	Device   string
	Mounted  bool
	hostPort string
	dev      hummingbird.Device
}

/*
func (bc *BirdCatcher) AllDevs() (devs []hummingbird.Device) {
	return bc.oring.AllDevices()
}
*/

func (bc *BirdCatcher) reconGetUnmounted(hostPort string, dataChan chan *ReconData, doneChan chan bool) {
	// TODO: log the errs
	defer func() {
		fmt.Println("mnmnmnmnm")
		doneChan <- true
	}()
	serverUrl := fmt.Sprintf("http://%s/recon/unmounted", hostPort)

	fmt.Println(serverUrl)
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", serverUrl, nil)
	if err != nil {
		bc.logger.Err(fmt.Sprintf("Could not create request to %s: %v", serverUrl, err))
		return
	}
	bc.logger.Info("lalalala qqqqqqqqqqqqqq")
	bc.logger.Err("lalalala rrrrrrrrrrrrrr")
	fmt.Println("777777777777777777777777777777777")
	resp, err := client.Do(req)
	if err != nil {
		//errs = append(errs, err)
		//serverToDev[serverUrl] = nil
		fmt.Println("333: ", err)
		bc.logger.Err(fmt.Sprintf("Could not do request to %s: %v", serverUrl, err))
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//errs = append(errs, err)
		fmt.Println("444", err)
		bc.logger.Err(fmt.Sprintf("Could not read resp to %s: %v", serverUrl, err))
		return
	}
	var serverReconData []*ReconData
	if err := json.Unmarshal(data, &serverReconData); err != nil {
		//errs = append(errs, err)
		bc.logger.Err(fmt.Sprintf("Could not parse json from %s: %v", serverUrl, err))
		fmt.Println("555", err)
		return
	}

	for _, rData := range serverReconData {
		rData.hostPort = hostPort
		dataChan <- rData
	}

}

func (bc *BirdCatcher) GatherReconData() (devs []*ReconData, errs []error) {

	allWeightedDevs := make(map[string]*hummingbird.Device)
	allServers := make(map[string]bool)

	for _, dev := range bc.oring.AllDevices() {
		if dev.Weight > 0 {
			allWeightedDevs[fmt.Sprintf(
				"%s:%d/%s", dev.Ip, dev.Port, dev.Device)] = &dev
			hostPort := fmt.Sprintf("%s:%d", dev.Ip, dev.Port)
			if _, ok := allServers[hostPort]; !ok {
				allServers[hostPort] = true
			}
		}
	}

	var allReconData []*ReconData
	serverCount := 0
	dataChan := make(chan *ReconData)
	doneChan := make(chan bool)

	for hostPort, _ := range allServers {
		go bc.reconGetUnmounted(hostPort, dataChan, doneChan)
		serverCount += 1
	}

	fmt.Println("the serverCnt: ", serverCount)
	for serverCount > 0 {
		select {
		case rd := <-dataChan:
			allReconData = append(allReconData, rd)
			delete(allWeightedDevs, fmt.Sprintf("%s/%s", rd.hostPort, rd.Device))
		case <-doneChan:
			serverCount -= 1
			fmt.Println("gotta done")
		}
	}

	for _, wDev := range allWeightedDevs {
		fmt.Println("666666")
		errs = append(errs, errors.New(fmt.Sprintf("%s:%d/%s was not found in recon", wDev.Ip, wDev.Port, wDev.Device)))
	}
	return allReconData, errs
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
	bc := &BirdCatcher{}
	bc.oring = objRing
	bc.logger = hummingbird.SetupLogger("LOG_LOCAL2", "birdcatcher", "") // fix at some point
	return bc, nil

}
