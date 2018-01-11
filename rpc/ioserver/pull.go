package ioserver

import (
	"io"

	pb "github.com/golang/protobuf/proto"

	"hpt.space/tapr"
	"hpt.space/tapr/errors"
	"hpt.space/tapr/log"
	"hpt.space/tapr/proto"
	"hpt.space/tapr/rpc"
)

func (s *server) PullPrepare(reqBytes []byte) (pb.Message, error) {
	_ = operation("pull/prepare")

	var req proto.PushPrepareRequest

	if err := pb.Unmarshal(reqBytes, &req); err != nil {
		return nil, err
	}

	tx := rpc.GenerateTx()

	log.Debug.Printf("rpc/ioserver[pull/prepare (tx: %s)]: %v", tx, req.Name)

	f, err := s.st.Open(tapr.PathName(req.Name))
	if err != nil {
		return nil, err
	}

	if req.Offset != 0 {
		if _, err := f.Seek(req.Offset, io.SeekStart); err != nil {
			return nil, err
		}
	}

	s.mu.fds[tx] = f

	return &proto.PushPrepareResponse{
		Tx: tx[:],
	}, nil
}

func (s *server) Pull(reqBytes []byte, done <-chan struct{}) (<-chan pb.Message, error) {
	op := operation("pull")

	var req proto.PullRequest

	if err := pb.Unmarshal(reqBytes, &req); err != nil {
		op.log(err)
		return nil, err
	}

	tx := rpc.MakeTx(req.Tx)

	log.Debug.Printf("rpc/ioserver[pull]: (tx: %s)", tx)

	f := s.mu.fds[tx]

	out := make(chan pb.Message)
	go func() {
		defer close(out)
		var cnk *proto.Chunk
		for {
			buf := make([]byte, 4096)
			n, err := f.Read(buf)
			if err != nil {
				cnk = &proto.Chunk{
					Error: errors.MarshalError(err),
				}

				delete(s.mu.fds, tx)
				return
			}

			cnk = &proto.Chunk{
				Data: buf[:n],
			}

			select {
			case out <- cnk:

			case <-done:
				log.Debug.Printf("rpc/ioserver[pull]: done closed; pull writer terminating")
				return
			}
		}
	}()

	return out, nil
}
