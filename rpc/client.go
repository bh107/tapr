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

package rpc // import "tapr.space/rpc"

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	pb "github.com/golang/protobuf/proto"

	"tapr.space"
	"tapr.space/errors"
	"tapr.space/flags"
	"tapr.space/log"
)

// Client is a partial tapr.Service that uses HTTP as a transport
// and implements authentication using out-of-band headers.
type Client interface {
	Close()

	// Invoke calls the given RPC method ("server/method") with the
	// given request message and decodes the response into the given
	// response message.
	Invoke(method string, req, resp pb.Message) error

	// Receive issues the request in req to the server and expects the
	// response to consist of a stream.
	Receive(method string, req pb.Message, stream StreamChan, done <-chan struct{}) error

	// Transmit issues a generic request to the server and streams the body.
	Transmit(method string, body io.Reader, resp pb.Message, done <-chan struct{}) error

	// Stream calls the given RPC streaming method. If req is nil, the request
	// body is set to body. If stream is non-nil, it will be used to decode the
	// response body.
	Stream(method string, req, resp pb.Message, body io.Reader, stream StreamChan, done <-chan struct{}) error
}

type httpClient struct {
	client *http.Client

	baseURL     string
	apiVersion  string
	targetStore string
}

// NewClient returns a new client that speaks to an HTTP server at a net
// address. The address is expected to be a raw network address with port
// number, as in domain.com:5580. The security level specifies the expected
// security guarantees of the connection. If proxyFor is an assigned endpoint,
// it indicates that this connection is being used to proxy request to that
// endpoint.
func NewClient(cfg tapr.Config, netAddr tapr.NetAddr) (Client, error) {
	const op = "rpc.NewClient"

	c := &httpClient{
		baseURL:     "http://" + string(netAddr),
		apiVersion:  "v1",
		targetStore: flags.Store,
	}

	t := &http.Transport{}

	c.client = &http.Client{Transport: t}

	return c, nil
}

// Invoke implements Client.
func (c *httpClient) Invoke(method string, req, resp pb.Message) error {
	const op = "rpc.Invoke"

	// Encode the payload directly.
	payload, err := pb.Marshal(req)
	if err != nil {
		return errors.E(op, err)
	}

	httpResp, err := c.invoke(op, method, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	// One-shot method, decode the response.
	err = readResponse(op, httpResp.Body, resp)
	if err != nil {
		return err
	}

	return nil
}

// Transmit implements Client.
func (c *httpClient) Transmit(method string, body io.Reader, resp pb.Message, done <-chan struct{}) error {
	const op = "rpc.Egress"

	httpResp, err := c.invoke(op, method, body)
	if err != nil {
		return err
	}

	// decode the response.
	err = readResponse(op, httpResp.Body, resp)
	if err != nil {
		return err
	}

	return nil
}

// Receive implements Client.
func (c *httpClient) Receive(method string, req pb.Message, stream StreamChan, done <-chan struct{}) error {
	const op = "rpc.Ingress"

	// Encode the payload directly.
	payload, err := pb.Marshal(req)
	if err != nil {
		return errors.E(op, err)
	}

	httpResp, err := c.invoke(op, method, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	go decodeStream(stream, httpResp.Body, done)

	return nil
}

// Stream implements Client.
func (c *httpClient) Stream(method string, req, resp pb.Message, body io.Reader, stream StreamChan, done <-chan struct{}) error {
	const op = "rpc.Stream"

	if (req == nil) == (body == nil) {
		return errors.E(op, errors.Str("exactly one of req and body must be nil"))
	}

	if (resp == nil) == (stream == nil) {
		return errors.E(op, errors.Str("exactly one of resp and stream must be nil"))
	}

	if req != nil {
		// Encode the payload directly.
		payload, err := pb.Marshal(req)
		if err != nil {
			return errors.E(op, err)
		}

		body = bytes.NewReader(payload)
	}

	httpResp, err := c.invoke(op, method, body)
	if err != nil {
		return err
	}

	if resp != nil {
		// decode the response.
		err = readResponse(op, httpResp.Body, resp)
		if err != nil {
			return err
		}
	}

	if stream != nil {
		go decodeStream(stream, httpResp.Body, done)
	}

	return nil
}

func (c *httpClient) makeRequest(op, method string, body io.Reader, header http.Header) (*http.Response, error) {
	header.Set("Content-Type", "application/octet-stream")

	// Make the HTTP request.
	url := fmt.Sprintf("%s/api/%s/%s/%s", c.baseURL, c.apiVersion, c.targetStore, method)
	httpReq, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, errors.E(op, errors.Invalid, err)
	}

	log.Debug.Printf("rpc/client: invoking %v", url)

	httpReq.Header = header
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, errors.E(op, errors.IO, err)
	}

	return resp, nil
}

func (c *httpClient) invoke(op, method string, body io.Reader) (resp *http.Response, err error) {
	resp, err = c.makeRequest(op, method, body, make(http.Header))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.Header.Get("Content-type") == "application/octet-stream" {
			return nil, errors.E(op, errors.UnmarshalError(msg))
		}

		return nil, errors.E(op, errors.IO, errors.Strf("%s: %s", resp.Status, msg))
	}

	return
}

func readResponse(op string, body io.ReadCloser, resp pb.Message) error {
	defer body.Close()

	respBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return errors.E(op, errors.IO, err)
	}

	if err := pb.Unmarshal(respBytes, resp); err != nil {
		return errors.E(op, errors.Invalid, err)
	}

	return nil
}

// Stubs for unused methods.
func (c *httpClient) Close() {}
