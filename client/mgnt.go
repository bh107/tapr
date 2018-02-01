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
