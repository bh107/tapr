package rpc

import (
	"encoding/binary"
	"io"

	pb "github.com/golang/protobuf/proto"

	"tapr.space/errors"
	"tapr.space/log"
	"tapr.space/proto"
)

// StreamChan describes a mechanism to report streamed messages to a client
// (the caller of Client.Invoke). Typically this interface should wrap a
// channel that carries decoded protocol buffers.
type StreamChan interface {
	// Send sends a proto-encoded message to the client.
	// If done is closed, the send should abort.
	Send(b []byte, done <-chan struct{}) error

	// Error sends an error condition to the client.
	Error(error)

	// Close closes the response channel.
	Close()
}

// decodeStream reads a stream of protobuf-encoded messages from r and sends
// them (without decoding them) to the given stream. If the done channel is
// closed then the stream and reader are closed and decodeStream returns.
func decodeStream(stream StreamChan, r io.ReadCloser, done <-chan struct{}) {
	defer stream.Close()
	defer r.Close()

	// A stream begins with the bytes "OK".
	var ok [2]byte
	if _, err := ReadFull(r, ok[:], done); err == io.ErrUnexpectedEOF {
		// Server closed the stream.
		return
	} else if err != nil {
		stream.Error(errors.E(errors.IO, err))
		return
	}

	if ok[0] != 'O' || ok[1] != 'K' {
		stream.Error(errors.E(errors.IO, errors.Str("unexpected stream preamble")))
	}

	// consume the rest of the stream
	ReadStream(r, stream, done)
}

// ReadStream reads all messages from r.
func ReadStream(r io.Reader, stream StreamChan, done <-chan struct{}) {
	var msgLen [4]byte
	var buf []byte

	for {
		// read the 4 byte, big-endian encoded int32
		if _, err := ReadFull(r, msgLen[:], done); err == io.ErrUnexpectedEOF {
			log.Debug.Print("rpc/ReadStream: unexpected EOF; stream done")
			return
		} else if err != nil {
			stream.Error(errors.E(errors.IO, err))
			return
		}

		l := binary.BigEndian.Uint32(msgLen[:])

		if cap(buf) < int(l) {
			buf = make([]byte, l)
		} else {
			buf = buf[:l]
		}

		if _, err := ReadFull(r, buf, done); err != nil {
			stream.Error(errors.E(errors.IO, err))
			return
		}

		if err := stream.Send(buf, done); err != nil {
			stream.Error(errors.E(errors.IO, err))
			return
		}
	}
}

// ReadMessage reads a single RPC message from r.
func ReadMessage(r io.Reader, done <-chan struct{}) ([]byte, error) {
	var msgLen [4]byte

	// read the 4 byte, big-endian encoded int32
	if _, err := ReadFull(r, msgLen[:], done); err != nil {
		return nil, err
	}

	l := binary.BigEndian.Uint32(msgLen[:])

	buf := make([]byte, l)

	// read message
	if _, err := ReadFull(r, buf, done); err != nil {
		return nil, err
	}

	return buf, nil
}

// ReadFull is like io.ReadFull but it will return io.EOF if the provided
// channel is closed.
func ReadFull(r io.Reader, b []byte, done <-chan struct{}) (int, error) {
	type result struct {
		n   int
		err error
	}

	ch := make(chan result, 1)

	go func() {
		// TODO(adg): this may leak goroutines if the requisite reads
		// never complete, but will that actually happen? It would be
		// great to have something like this hooked into the runtime.
		n, err := io.ReadFull(r, b)
		ch <- result{n, err}
	}()

	select {
	case r := <-ch:
		return r.n, r.err
	case <-done:
		return 0, io.EOF
	}
}

// ChunkStream is a channel of proto.Chunk.
type ChunkStream chan proto.Chunk

// Send implements StreamChan.
func (s ChunkStream) Send(b []byte, done <-chan struct{}) error {
	var cnk proto.Chunk
	if err := pb.Unmarshal(b, &cnk); err != nil {
		return err
	}

	select {
	case s <- cnk:
	case <-done:
	}

	return nil
}

// Close implements StreamChan.
func (s ChunkStream) Close() {
	log.Debug.Print("rpc/ChunkStream.Close: closing")
	close(s)
}

// Error implements StreamChan.
func (s ChunkStream) Error(err error) {
	log.Debug.Printf("rpc/ChunkStream.Error: %v", err)
	s <- proto.Chunk{Error: errors.MarshalError(err)}
}

// LogStream is an implementation of StreamChan carrying proto.Chunk.
type LogStream chan proto.PushLogEntry

// Send implements StreamChan.
func (s LogStream) Send(b []byte, done <-chan struct{}) error {
	var ack proto.PushLogEntry
	if err := pb.Unmarshal(b, &ack); err != nil {
		return err
	}

	select {
	case s <- ack:
		log.Debug.Printf("sent something...?")
	case <-done:
	}

	return nil
}

// Close implements StreamChan.
func (s LogStream) Close() {
	log.Debug.Print("rpc/LogStream.Close: closing")
	close(s)
}

// Error implements StreamChan.
func (s LogStream) Error(err error) {
	log.Debug.Printf("rpc/LogStream.Error: %v", err)
	s <- proto.PushLogEntry{Error: errors.MarshalError(err)}
}
