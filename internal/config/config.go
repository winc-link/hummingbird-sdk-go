/*******************************************************************************
 * Copyright 2017.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package config

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

const (
	DefaultConfigFilePath     = "/etc/driver/res/configuration.toml"
	TestDefaultConfigFilePath = "cmd/res/configuration.toml"

	Core       = "Core"
	MQTTBroker = "MQTTBroker"
)

type (
	LogConfig struct {
		FileName string
		LogLevel string
	}

	RPCConfig struct {
		Address  string
		UseTLS   bool
		CertFile string
		KeyFile  string
	}

	ClientInfo struct {
		Address string
		// 是否启用tls
		UseTLS bool
		// ca cert
		CertFilePath string
		// mqtt clientId
		ClientId string
		// mqtt username
		Username string
		// mqtt password
		Password string
	}

	ServiceInfo struct {
		ID     string
		Name   string
		Server RPCConfig
		GwId   string
	}

	DriverConfig struct {
		Logger      LogConfig
		Clients     map[string]ClientInfo
		Service     ServiceInfo
		CustomParam string
	}
)

func (d *DriverConfig) GetServiceID() string {
	return d.Service.ID
}

var FilePath string

func (d *DriverConfig) ValidateConfig() error {
	if len(d.Clients[Core].Address) == 0 {
		return fmt.Errorf("resource address client not configured")
	}
	return nil
}

func ParseConfig() (*DriverConfig, error) {
	var (
		err      error
		contents []byte
		dc       DriverConfig
	)
	if contents, err = ioutil.ReadFile(FilePath); err != nil {
		return nil, errors.New(fmt.Sprintf("service not load configuration file: %s", err.Error()))
	}
	if err = toml.Unmarshal(contents, &dc); err != nil {
		return nil, errors.New(fmt.Sprintf("service not load configuration file: %s", err.Error()))
	}
	return &dc, nil
}
