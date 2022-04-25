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

package cni

import (
	typesVer "github.com/containernetworking/cni/pkg/types/100"
	"github.com/yametech/global-ipam/pkg/config"
)

const GLOBAL_IPAM = "global-ipam"

const UNIX_SOCK_PATH = "/var/run/global-ipam.sock"

type Cni interface {
	Allocate(namespace, pod string, cfg *config.Net) (*typesVer.Result, error)
	Release(namespace, pod string) error
}

type AllocateResponse struct {
	Reserved bool             `json:"reserved"`
	Result   *typesVer.Result `json:"result"`
	Error    error            `json:"error"`
}

type ReleaseResponse struct {
	IsRelease bool  `json:"isRelease"`
	Error     error `json:"error"`
}
