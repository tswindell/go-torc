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

/*

  Package torc provides an API for communicating with Tor via a control socket
  connection. This package was developed using the "Tor control protocol
  (Version 1)" specification, from here:

    https://gitweb.torproject.org/torspec.git/plain/control-spec.txt

  The API was designed to be as closely mapped to the control spec as possible,
  so it may seem a little mechanical. This choice was made to ensure that less
  effort is required to modify the implementation as the protocol is modified.

  The main entry point to this API is the Controller type, it handles
  connecting, disconnecting, authentication, sending & receiving messages,
  incoming signal handling and command invokation. Currently only a subset of
  commands are implemented, more will hopefully be implemented as this package
  grows older.

*/
package torc

import (
    "fmt"
    "net"
    "os"
    "time"
    "reflect"
)

// Function template for dialer parameter.
type DialerFunc func(string, string) (net.Conn, error)

// The Controller type handles connecting, disconnecting, authenticating,
// sending & receiving of messages, event dispatching and command invokation.
// You may supply a custom dialer function for connecting to the control socket
// through a proxy, or some other connection means by replacing the DialerFunc
// ``Dialer'' property before calling Connect().
//
// Once a connection is established you may use the Controller Command API to
// send command requests and receive responses. Authentication is handled
// automatically, though when a Password is required you must specify one before
// you call the Connect() method.
//
type Controller struct {
    // The function to use for dialing remote connection.
    Dialer   DialerFunc
    network  string
    hostport string

    // Control socket connection instance.
    connection *net.Conn

    isConnected bool

    // Incoming response message queue.
    in chan ResponseBuffer

    // Incoming message parser instance.
    parser *Parser

    // Optional password to use during authentication.
    Password        string

    authenticator   Authenticator
    isAuthenticated bool
}

// Creates a new Controller instance, for connecting to a Tor service's
// control socket through the specified dialer network and hostport
// parameters.
func NewController(network, hostport string) *Controller {
    c := new(Controller)

    if len(os.Getenv("TORC_LOG_COMMS")) > 0 {
       COMMS_LOGGING = true
    }

    c.Dialer   = net.Dial
    c.network  = network
    c.hostport = hostport

    return c
}

// Connect this controller to the control socket of a Tor service by using
// this instances Dialer function to initiate a connection. Once connected
// automated authentication takes over. If no error occurs during this process
// then you may begin to perform API requests.
func (c *Controller) Connect() error {
    if c.IsConnected() {
        LogInfo("Attempt to dial when already connected, failing silently.")
        return nil
    }

    LogInfo("Attemping to dial remote: (%s) %s", c.network, c.hostport)
    conn, e := c.Dialer(c.network, c.hostport)
    if e != nil {
        LogInfo("Failed to connect to remote: %v", e)
        return e
    }

    LogInfo("Connection established.")
    c.connection = &conn
    c.isConnected = true

    c.in = make(chan ResponseBuffer, 1)
    c.parser = NewParser(conn, c.in)

    // Kickstart reader/parser goroutine.
    LogInfo("Starting reader.")
    go c.parser.Run()

    // Send PROTOCOLINFO request to get authentication mechanisms.
    protoinfo, e := c.ProtocolInfo()
    if e != nil {
        LogInfo("PROTOCOLINFO request failed: %v", e)
        return e
    }

    // TODO: Expose an API for managing user preferences?
    authPrefs := []Authenticator{
                     &CookieAuthenticator{},
                     &PasswordAuthenticator{},
                     &OpenAuthenticator{},
                 }

    // Automatically Select prefered authentication method.
    for _, i := range authPrefs {
        for _, v := range protoinfo.AuthMethods() {
            if i.MethodName() != v { continue }
            c.authenticator = i
            break
        }
    }

    if c.authenticator == nil {
        return fmt.Errorf("Failed to find compatible authentication method.")
    }

    if e := c.authenticator.Authenticate(c, protoinfo); e != nil {
        return fmt.Errorf("Authentication failed!")
    }

    LogInfo("Successfully authenticated controller.")
    return nil
}

// Close this Controller instances connection to Tor service.
func (c *Controller) Close() {
    if c.connection == nil {
        return
    }

    (*c.connection).Close()

    close(c.in)

    c.connection = nil
    c.isConnected = false
    c.isAuthenticated = false
}

// Returns true if Controller instance believes it's connected.
func (c *Controller) IsConnected() bool {
    return c.isConnected
}

// Returns true if Controller instance has been authenticated.
func (c *Controller) IsAuthenticated() bool {
    return c.isAuthenticated
}

// Send message through control socket.
func (c *Controller) SendMessage(buffer LineBuffer) error {
    LogComms("<<", buffer)
    _, e := (*c.connection).Write(buffer.Normalize())
    return e
}

// Send request through control socket, and populate response with reply.
func (c *Controller) Request(request ControlRequest, response interface{}) error {
    e := c.SendMessage(request.Serialize())
    if e != nil {
        LogError("Failed to send request: %v", e)
        return e
    }

    // Wait for reply.
    select {
        case buff := <-c.in:
            r := NewResponse(request, buff)
            // VOODOO FOR SETTING BASE INSTANCE
            v := reflect.ValueOf(response).Elem()
            v.Field(0).Set(reflect.ValueOf(r))
            return nil

        case <-time.After(request.ResponseTimeout()):
            LogWarn("Timeout waiting for reply.")
    }

    return fmt.Errorf("Timeout waiting for reply.")
}

