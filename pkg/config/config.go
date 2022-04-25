// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/yametech/global-ipam/pkg/allocator"
)

type Ipam struct {
	Type   string               `json:"type"`
	Ranges []allocator.RangeSet `json:"ranges"`
	Routes []*types.Route       `json:"routes"`
}
type Net struct {
	CniVersion string `json:"cniVersion"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Master     string `json:"master"`
	ResolvConf string `json:"resolvConf"`
	Namespace  string `json:"namespace"`
	PodName    string `json:"podName"`
	Ipam       *Ipam  `json:"ipam"`
	Args       *struct {
		Cni *struct {
			IPs []net.IP `json:"ips"`
		} `json:"cni"`
	} `json:"args"`
}

func LoadConfig(bytes []byte, envArgs string) (*Net, string, error) {
	n := Net{}
	if err := json.Unmarshal(bytes, &n); err != nil {
		return nil, "", fmt.Errorf("failed to load netconf: %v", err)
	}

	if n.Ipam == nil {
		return nil, "", fmt.Errorf("IPAM config missing 'ipam' key")
	}

	podName, err := parseValueFromArgs("K8S_POD_NAME", envArgs)
	if err != nil {
		return nil, "", err
	}

	n.PodName = podName

	podNamespace, err := parseValueFromArgs("K8S_POD_NAMESPACE", envArgs)
	if err != nil {
		return nil, "", err
	}
	n.Namespace = podNamespace

	return &n, n.CniVersion, nil

}

func parseValueFromArgs(key, argString string) (string, error) {
	if argString == "" {
		return "", errors.New("CNI_ARGS is required")
	}
	args := strings.Split(argString, ";")
	for _, arg := range args {
		if strings.HasPrefix(arg, fmt.Sprintf("%s=", key)) {
			value := strings.TrimPrefix(arg, fmt.Sprintf("%s=", key))
			if len(value) > 0 {
				return value, nil
			}
		}
	}
	return "", fmt.Errorf("%s is required in CNI_ARGS", key)
}
