// Copyright 2018 Klaus Birkelund Abildgaard Jensen
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

package client

import (
	"tapr.space"
	"tapr.space/mgnt"
	"tapr.space/rpc"
	"tapr.space/store/tape"
	"tapr.space/store/tape/proto"
)

// ManagementClient implements mgnt.Client
type ManagementClient struct {
	config tapr.Config
	client rpc.Client
}

var _ mgnt.Client = (*ManagementClient)(nil)

// NewManagementClient creates a Client that uses the given
// configuration to access the various Tapr servers.
func NewManagementClient(config tapr.Config) mgnt.Client {
	m := &ManagementClient{config: config}

	client, err := rpc.NewClient(config, "localhost:8080")
	if err != nil {
		return nil
	}

	m.client = client

	return m
}

// Volumes implements mgnt.Client.
func (m *ManagementClient) Volumes() ([]tape.Volume, error) {
	var resp proto.StatusResponse
	if err := m.client.Invoke("inv/volumes", &proto.StatusRequest{}, &resp); err != nil {
		return nil, err
	}

	return proto.TaprVolumes(resp.Volumes), nil
}
