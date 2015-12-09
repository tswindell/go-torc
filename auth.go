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
    "fmt"
    "io/ioutil"
)

type AuthResponse struct { *BaseControlResponse }

// The Authenticator interface defines the API for plugin authentication modules
// to implement. When adding authentication modules, make sure they're
// registered in the Controller before application calls "Connect()"
type Authenticator interface {
    MethodName() string
    Authenticate(*Controller, *ProtocolInfoResponse) error
}


// An OpenAuthenticator implements the Authenticator interface to provide an
// "auth-less" authentication to the control socket.
type OpenAuthenticator struct {}

func (a *OpenAuthenticator) MethodName() string { return "NULL" }

func (a *OpenAuthenticator) Authenticate(c *Controller, protoinfo *ProtocolInfoResponse) error {
    LogInfo("Attempting open authentication...")

    response := &AuthResponse{}
    e := c.Request(NewRequest(COMMAND_AUTHENTICATE), response)
    if e != nil {
        LogWarn("Failed to send request: %v", e)
        return e
    }

    c.isAuthenticated = response.IsSuccess()
    if !c.IsAuthenticated() {
        return fmt.Errorf("Authentication failed!")
    }

    return nil
}

// A CookieAuthenticator implements the Authenticator interface to provide the
// cookie authentication to the control socket.
type CookieAuthenticator struct {}

func (a *CookieAuthenticator) MethodName() string { return "COOKIE" }

func (a *CookieAuthenticator) Authenticate(c *Controller, protoinfo *ProtocolInfoResponse) error {
    LogInfo("Attempting cookie authentication...")

    cookie, e := ioutil.ReadFile(protoinfo.AuthCookieFile())
    if e != nil {
        LogError("Failed to read cookie: %v", e)
        return e
    }

    response := &AuthResponse{}
    e = c.Request(NewRequest(fmt.Sprintf("AUTHENTICATE %x", cookie)), response)
    if e != nil || !response.IsSuccess() {
        LogWarn("Failed to send request: %v", e)
        return e
    }

    c.isAuthenticated = response.IsSuccess()
    if !c.IsAuthenticated() {
        return fmt.Errorf("Authentication failed!")
    }

    return nil
}

// A PasswordAuthenticator implements the Authenticator interface to provide
// password based authentication to the control socket.
type PasswordAuthenticator struct { }

func (a *PasswordAuthenticator) MethodName() string { return "HASHEDPASSWORD" }

func (a *PasswordAuthenticator) Authenticate(c *Controller, protoinfo *ProtocolInfoResponse) error {
    LogInfo("Attempting password authentication...")

    response := &AuthResponse{}
    e := c.Request(NewRequest(fmt.Sprintf("AUTHENTICATE \"%s\"", c.Password)), response)
    if e != nil || !response.IsSuccess() {
        LogWarn("Failed to send request: %v", e)
        return e
    }

    c.isAuthenticated = response.IsSuccess()
    if !c.IsAuthenticated() {
        return fmt.Errorf("Authentication failed!")
    }

    return nil
}

// A SafeCookieAuthenticator implements the Authenticator interface to provide
// a more secure form of cookie based authentication, where the cookie data is
// not transmitted in plain text.
type SafeCookieAuthenticator struct {
    CookieFile string
}

func (a *SafeCookieAuthenticator) MethodName() string { return "SAFECOOKIE" }

func (a *SafeCookieAuthenticator) Authenticate(c *Controller, protoinfo *ProtocolInfoResponse) error {
    LogInfo("Attempting safe-cookie authentication...")
    return NotImplemented()
}

