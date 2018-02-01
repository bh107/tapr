// Package client implements a tapr.Client.
package client // import "tapr.space/client"

import (
	"encoding/binary"
	"io"

	pb "github.com/golang/protobuf/proto"

	"tapr.space"
	"tapr.space/errors"
	"tapr.space/log"
	"tapr.space/proto"
	"tapr.space/rpc"
)

// Client implements tapr.Client.
type Client struct {
	config tapr.Config
	client rpc.Client
}

var _ tapr.Client = (*Client)(nil)

// New creates a Client that uses the given configuration to
// access the various Tapr servers.
func New(config tapr.Config) tapr.Client {
	cl := &Client{config: config}

	client, err := rpc.NewClient(config, "localhost:8080")
	if err != nil {
		return nil
	}

	cl.client = client

	return cl
}

// Stat implements tapr.Client.
func (c *Client) Stat(name tapr.PathName) (*tapr.FileInfo, error) {
	statReq := &proto.StatRequest{
		Name: string(name),
	}

	var statResp proto.StatResponse
	if err := c.client.Invoke("io/stat", statReq, &statResp); err != nil {
		return nil, err
	}

	log.Debug.Print("client: stat ok")

	return &tapr.FileInfo{Size: statResp.Size}, nil
}

// Pull implements tapr.Client.
func (c *Client) Pull(name tapr.PathName, w io.Writer) error {
	return c.PullFile(name, w, 0 /* offset */)
}

// PullFile implements tapr.Client.
func (c *Client) PullFile(name tapr.PathName, w io.Writer, offset int64) error {
	prepareReq := &proto.PullPrepareRequest{
		Name:   string(name),
		Offset: offset,
	}

	var prepareResp proto.PullPrepareResponse
	if err := c.client.Invoke("io/pull/prepare", prepareReq, &prepareResp); err != nil {
		return err
	}

	tx := rpc.MakeTx(prepareResp.Tx)

	log.Debug.Printf("client: pull/prepare ok (tx: %v)", tx.String())

	done := make(chan struct{})
	stream := make(rpc.ChunkStream)

	pullReq := &proto.PullRequest{Tx: prepareResp.Tx}

	// setup the stream
	if err := c.client.Receive("io/pull", pullReq, stream, done); err != nil {
		return err
	}

	// process the received chunks
	for cnk := range stream {
		if _, err := w.Write(cnk.Data); err != nil {
			log.Debug.Printf("client.PullFile: could not write: %v", err)
			return err
		}

		log.Debug.Printf("client.PullFile: received %d bytes", len(cnk.Data))
	}

	return nil
}

// Append implements tapr.Client.
func (c *Client) Append(name tapr.PathName, rd io.Reader) error {
	return c.PushFile(name, rd, true /* append */)
}

// Push implements tapr.Client.
func (c *Client) Push(name tapr.PathName, rd io.Reader) error {
	return c.PushFile(name, rd, false /* append */)
}

// PushFile implements tapr.Client.
func (c *Client) PushFile(name tapr.PathName, rd io.Reader, append bool) error {
	prepareReq := &proto.PushPrepareRequest{
		Name:   string(name),
		Append: append,
	}

	var prepareResp proto.PushPrepareResponse
	if err := c.client.Invoke("io/push/prepare", prepareReq, &prepareResp); err != nil {
		return err
	}

	tx := rpc.MakeTx(prepareResp.Tx)

	log.Debug.Printf("client.Push: prepare ok (tx: %s)", tx)

	done := make(chan struct{})

	pr, pw := io.Pipe()

	go func() {
		var pushResp proto.PushResponse
		c.client.Transmit("io/push", pr, &pushResp, done)
		if pushResp.Error != nil {
			log.Debug.Printf("client.Push: error: %v", errors.UnmarshalError(pushResp.Error))
			return
		}

		log.Debug.Printf("client.Push: push done")
	}()

	// write the transaction identifier
	if _, err := pw.Write(tx[:]); err != nil {
		log.Debug.Printf("client.Push: could not write to pipe: %v", err)
		return err
	}

	go func() {
		var lenBytes [4]byte // stores a uint32, the length of each output message
		buf := make([]byte, 4096)
		for {
			n, err := rd.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Debug.Printf("client.Push: EOF reached, writer shutting down")
					close(done)
					break
				} else if err != nil {
					log.Error.Printf("client.Push: %v", err)
					break
				}
			}

			b, err := pb.Marshal(&proto.Chunk{
				Data: buf[:n],
			})
			if err != nil {
				log.Error.Printf("client.Push: error marshalling proto: %v", err)
				break
			}

			binary.BigEndian.PutUint32(lenBytes[:], uint32(len(b)))

			if _, err := pw.Write(lenBytes[:]); err != nil {
				log.Debug.Printf("client.Push: could not write to pipe: %v", err)
				break
			}

			if _, err := pw.Write(b); err != nil {
				log.Debug.Printf("client.Push: could not write to pipe: %v", err)
				break
			}
		}
	}()

	stream := make(rpc.LogStream)

	logRequest := &proto.PushLogRequest{Tx: prepareResp.Tx}

	if err := c.client.Receive("io/push/log", logRequest, stream, done); err != nil {
		return err
	}

	for entry := range stream {
		if entry.Error != nil {
			log.Debug.Printf("client.Push: error: %v", errors.UnmarshalError(entry.Error))
			continue
		}

		log.Debug.Printf("client.Push: log received: %v", entry.Seq)

	}

	return nil
}
