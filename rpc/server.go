// Copyright 2016 The Upspin Authors. All rights reserved.
// Copyright 2017 The Tapr Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
//    * Redistributions of source code must retain the above copyright
//      notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
//      copyright notice, this list of conditions and the following
//      disclaimer in the documentation and/or other materials provided
//      with the distribution.
//    * Neither the name of Google Inc. nor the names of its
//      contributors may be used to endorse or promote products derived
//      from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package rpc // import "hpt.space/tapr/rpc"

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	pb "github.com/golang/protobuf/proto"

	"hpt.space/tapr"
	"hpt.space/tapr/errors"
	"hpt.space/tapr/log"
)

// Method describes an authenticated RPC method.
type Method func(reqBytes []byte) (pb.Message, error)

// Egress describes a streaming RPC method.
type Egress func(reqBytes []byte, done <-chan struct{}) (<-chan pb.Message, error)

// Ingress describes a streaming RPC method.
type Ingress func(body io.Reader, done <-chan struct{}) (pb.Message, error)

// Service describes an RPC service.
type Service struct {
	// The name of the service, which forms the first path component of any
	// HTTP request.
	Name string

	// The RPC methods to serve.
	Methods map[string]Method

	// The streaming RPC methods to serve.
	Egress map[string]Egress

	Ingress map[string]Ingress
}

// Tx is a state.
type Tx [20]byte

func (tx Tx) String() string {
	return fmt.Sprintf("%x...", tx[:4])
}

// GenerateTx generates a random token of the specified size.
func GenerateTx() Tx {
	var tx Tx
	rand.Read(tx[:])

	return tx
}

// MakeTx creates a Tx from a byte slice.
func MakeTx(p []byte) (tx Tx) {
	if len(p) != len(tx) {
		panic("invalid byte slice")
	}

	copy(tx[:], p)

	return
}

// ParseTx parses a string representation of a transaction token into a
// token type.
func ParseTx(s string) (Tx, error) {
	var tx Tx

	decoded, err := hex.DecodeString(s)
	if err != nil {
		return tx, err
	}

	if len(decoded) != len(tx) {
		return tx, errors.E(errors.Invalid)
	}

	copy(tx[:], decoded)

	return tx, nil
}

// NewServer returns a new Server that uses the given config.
func NewServer(cfg tapr.Config, svc Service) http.Handler {
	if svc.Name == "" {
		panic("config provided with empty Name")
	}

	return &serverImpl{
		config:  cfg,
		service: svc,
	}
}

type serverImpl struct {
	config  tapr.Config
	service Service
}

// ServeHTTP exposes the configured Service as an HTTP API.
func (s *serverImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d := &s.service
	prefix := "/api/v1/" + d.Name + "/"

	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.NotFound(w, r)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, prefix)

	method := d.Methods[name]
	egress := d.Egress[name]
	ingress := d.Ingress[name]

	if method == nil && egress == nil && ingress == nil {
		http.NotFound(w, r)
		return
	}

	switch {
	case method != nil:
		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := method(body)
		sendResponse(w, resp, err)

	case egress != nil:
		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		serveEgress(egress, w, body)

	case ingress != nil:
		serveIngress(ingress, w, r.Body)

	default:
		panic("this should never happen")
	}
}

func serveIngress(s Ingress, w http.ResponseWriter, body io.Reader) {
	done := make(chan struct{})

	connClosed := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-connClosed
		close(done)
	}()

	resp, err := s(body, done)

	log.Debug.Print("rpc.serveIngress: completed; sending response")
	sendResponse(w, resp, err)
}

func serveEgress(s Egress, w http.ResponseWriter, body []byte) {
	done := make(chan struct{})
	msgs, err := s(body, done)
	if err != nil {
		sendError(w, err)
		return
	}

	connClosed := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-connClosed
		close(done)
	}()

	// Write the headers, beginning the stream.
	w.Write([]byte("OK"))
	w.(http.Flusher).Flush()

	var lenBytes [4]byte // stores a uint32, the length of each output message
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return
			}

			if done == nil {
				// Drop this message as there's nobody to deliver to.
				continue
			}

			b, err := pb.Marshal(msg)
			if err != nil {
				log.Error.Printf("rpc/stream: error encoding proto in stream: %v", err)
				return
			}

			binary.BigEndian.PutUint32(lenBytes[:], uint32(len(b)))
			if _, err := w.Write(lenBytes[:]); err != nil {
				return
			}
			if _, err := w.Write(b); err != nil {
				return
			}
			w.(http.Flusher).Flush()

		case <-done:
			done = nil
		}
	}
}

func sendResponse(w http.ResponseWriter, resp pb.Message, err error) {
	if err != nil {
		sendError(w, err)
		return
	}

	payload, err := pb.Marshal(resp)
	if err != nil {
		log.Error.Printf("error encoding response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(payload)
}

func sendError(w http.ResponseWriter, err error) {
	if _, ok := err.(*errors.Error); !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h := w.Header()

	h.Set("Content-type", "application/octet-stream")
	w.WriteHeader(http.StatusInternalServerError)

	w.Write(errors.MarshalError(err))
}
