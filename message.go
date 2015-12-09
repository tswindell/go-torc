/*
 * Copyright (c) 2015 Tom Swindell (t.swindell@rubyx.co.uk)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package torc

import (
    "time"
)

// The Message interface describes the interface used to serialize outbound
// control messages.
type Message interface {
    Serialize() LineBuffer
}

// The ControlRequest interface describes the interface used for outbound
// control command requests.
type ControlRequest interface {
    ResponseTimeout() time.Duration
    Serialize() LineBuffer
}

// The ControlResponse interface describes the interface used for incoming
// control command responses.
type ControlResponse interface {
    Request() ControlRequest

    Status() int
    StatusText() string

    IsSuccess() bool
}

// The BaseControlRequest type is the base type of all command request tpes.
type BaseControlRequest struct {
    buffer  LineBuffer
    timeout time.Duration
}

// Instantiates a new BaseControlRequest instance.
func NewRequest(data string) *BaseControlRequest {
    m := new(BaseControlRequest)
    m.buffer = make(LineBuffer, 0)
    m.buffer = append(m.buffer, data)
    m.timeout = time.Second * 5
    return m
}

func (m *BaseControlRequest) ResponseTimeout() time.Duration {
    return m.timeout
}

func (m *BaseControlRequest) Serialize() LineBuffer {
    return m.buffer
}

// The BaseControlResponse type is the base type for all command response types.
type BaseControlResponse struct {
    Request ControlRequest

    Buffer  ResponseBuffer
}

// Instantiates a new BaseControlResponse object.
func NewResponse(req ControlRequest, data ResponseBuffer) *BaseControlResponse {
    m := new(BaseControlResponse)
    m.Buffer  = data
    m.Request = req
    return m
}

/*
  According to:

    https://gitweb.torproject.org/torspec.git/tree/control-spec.txt

  The TC protocol currently uses the following first characters:
  2yz   Positive Completion Reply
  The command was successful; a new request can be started.

  4yz   Temporary Negative Completion reply
  The command was unsuccessful but might be reattempted later.

  5yz   Permanent Negative Completion Reply
  The command was unsuccessful; the client should not try exactly
  that sequence of commands again.

  6yz   Asynchronous Reply
  Sent out-of-order in response to an earlier SETEVENTS command.

  The following second characters are used:

  x0z   Syntax
  Sent in response to ill-formed or nonsensical commands.

  x1z   Protocol
  Refers to operations of the Tor Control protocol.

  x5z   Tor
  Refers to actual operations of Tor system.
*/

// Returns the numerical status code of the response as mentioned in the
// EndReplyLine
func (r *BaseControlResponse) Status() int {
    return r.Buffer.EndReplyLine.Status()
}

// Returns the text portion of the EndReplyLine of the response.
func (r *BaseControlResponse) StatusText() string {
    return r.Buffer.EndReplyLine.StatusText()
}

// Returns a boolean indicating whether the request was successful, this is an
// abstraction over the Status() method.
func (r *BaseControlResponse) IsSuccess() bool {
    status := r.Status()
    if status > 299 && status < 600 {
        return false
    }
    return true
}

