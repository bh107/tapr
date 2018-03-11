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

package invserver // import "tapr.space/rpc/invserver"

import (
	"fmt"
	"net/http"

	pb "github.com/golang/protobuf/proto"

	"tapr.space"
	"tapr.space/errors"
	"tapr.space/log"
	"tapr.space/rpc"
	"tapr.space/store/tape/inv"
	"tapr.space/store/tape/proto"
)

type server struct {
	config tapr.Config

	inv inv.Inventory
}

// New returns a new http.Handler that presents an inventory server
// as a service.
func New(cfg tapr.Config, inv inv.Inventory) http.Handler {
	s := &server{
		config: cfg,
		inv:    inv,
	}

	return rpc.NewServer(cfg, rpc.Service{
		Name: "inv",
		Methods: map[string]rpc.Method{
			"volumes": s.Volumes,
		},
	})
}

func (s *server) Volumes(reqBytes []byte) (pb.Message, error) {
	op := operation("status")

	vols, err := s.inv.Volumes()
	if err != nil {
		op.log(err)
		return &proto.StatusResponse{Error: errors.MarshalError(err)}, nil
	}

	resp := &proto.StatusResponse{
		Volumes: proto.VolumeProtos(vols),
	}

	return resp, nil
}

func logf(format string, args ...interface{}) operation {
	s := fmt.Sprintf(format, args...)
	log.Debug.Print("rpc/invserver: " + s)
	return operation(s)
}

type operation string

func (op operation) log(err error) {
	logf("%v failed: %v", op, err)
}
