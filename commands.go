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
    "regexp"
    "strconv"
    "strings"
)

// Tor control protocol command constants, use these when building commands from
// scratch using NewRequest. Currently only a subset of these commands are
// implemented in the API, hopefully more will be added as torc matures..
const (
              COMMAND_SETCONF = "SETCONF"
            COMMAND_RESETCONF = "RESETCONF"
              COMMAND_GETCONF = "GETCONF"
            COMMAND_SETEVENTS = "SETEVENTS"
         COMMAND_AUTHENTICATE = "AUTHENTICATE"
             COMMAND_SAVECONF = "SAVECONF"
               COMMAND_SIGNAL = "SIGNAL"
           COMMAND_MAPADDRESS = "MAPADDRESS"
              COMMAND_GETINFO = "GETINFO"
        COMMAND_EXTENDCIRCUIT = "EXTENDCIRCUIT"
    COMMAND_SETCIRCUITPURPOSE = "SETCIRCUITPURPOSE"
     COMMAND_SETROUTERPURPOSE = "SETROUTEPURPOSE"
         COMMAND_ATTACHSTREAM = "ATTACHSTREAM"
       COMMAND_POSTDESCRIPTOR = "POSTDESCRIPTOR"
       COMMAND_REDIRECTSTREAM = "REDIRECTSTREAM"
          COMMAND_CLOSESTREAM = "CLOSESTREAM"
         COMMAND_CLOSECIRCUIT = "CLOSECIRCUIT"
                 COMMAND_QUIT = "QUIT"
           COMMAND_USEFEATURE = "USEFEATURE"
              COMMAND_RESOLVE = "RESOLVE"
         COMMAND_PROTOCOLINFO = "PROTOCOLINFO"
             COMMAND_LOADCONF = "LOADCONF"
        COMMAND_TAKEOWNERSHIP = "TAKEOWNERSHIP"
        COMMAND_AUTHCHALLENGE = "AUTHCHALLENGE"
           COMMAND_DROPGUARDS = "DROPGUARDS"
              COMMAND_HSFETCH = "HSFETCH"
            COMMAND_ADD_ONION = "ADD_ONION"
            COMMAND_DEL_ONION = "DEL_ONION"
               COMMAND_HSPOST = "HSPOST"
) // TODO: Wrap in a ControlCommand semantic type.

// The GetInfoResponse type is returned by the GetInfo command method.
type GetInfoResponse struct { *BaseControlResponse }

// Return first MidReply value, or first DataReply value
func (m *GetInfoResponse) Value() string {
    // Try to get value from MidReplyLine and return.
    if len(m.Buffer.MidReplyLines) == 1 {
        return strings.SplitN(string(m.Buffer.MidReplyLines[0]), "=", 2)[1]
    }
    // Otherwise get value from DataReply lines and return.
    if len(m.Buffer.DataReplyLines) == 1 {
        data := m.Buffer.DataReplyLines[0]
        return strings.Join([]string(data)[1:], "\n")
    }

    //FIXME: Do we really want this function to return empty if there are more
    // than one mid or data reply lines? Maybe we should just return the first
    // one.
    return ""
}

// Return specific mid or data reply value relating to key.
func (m *GetInfoResponse) ValueOf(key string) string {
    // Try to get value from a MidReplyLine and return it.
    text := __find_prefix_mrl(m.Buffer.MidReplyLines, key)
    if text != "" { return strings.TrimPrefix(text, key + "=") }
    // Otherwise get value from a DataReply and return it (or empty string)
    return __find_prefix_drl(m.Buffer.DataReplyLines, key)
}

// Return all received values as a map.
func (m *GetInfoResponse) ValueAll() map[string]string {
    results := make(map[string]string)
    // Iterate over mid reply lines and add to dictionary.
    for _, v := range m.Buffer.MidReplyLines {
        parts := strings.SplitN(v.Text(), "=", 2)
        results[parts[0]] = parts[1]
    }
    // Iterate over data reply lines and add data to dictionary.
    for _, v := range m.Buffer.DataReplyLines {
        k := strings.TrimSuffix(v.Text(), "=")
        v := strings.Join(v[1:], "\n")
        results[k] = v
    }
    return results
}

// Perform GETINFO command request. Returns GetInfoResponse instance reflecting
// command result.
func (c *Controller) GetInfo(keys []string) (*GetInfoResponse, error) {
    request  := NewRequest(COMMAND_GETINFO + " " + strings.Join(keys, " "))
    response := &GetInfoResponse{}
    return response, c.Request(request, response)
}

// The ProtocolInfoResponse type is returned by the ProtocolInfo command method.
type ProtocolInfoResponse struct { *BaseControlResponse }

// Get protocol version number integer from response.
func (m *ProtocolInfoResponse) Protocol() int {
    text := __find_prefix_mrl(m.Buffer.MidReplyLines, "PROTOCOLINFO")
    i, e := strconv.Atoi(strings.SplitN(text, " ", 2)[1])
    if e != nil { return -1 }
    return i
}

// Get protocol auth line key value pairs in a map.
func (m *ProtocolInfoResponse) Auth() map[string]string {
    text := __find_prefix_mrl(m.Buffer.MidReplyLines, "AUTH")
    return __make_variable_map(text[len("AUTH"):])
}

// Get auth methods from protocol auth line as strings.
func (m *ProtocolInfoResponse) AuthMethods() []string {
    return strings.SplitN(m.Auth()["METHODS"], ",", -1)
}

// Get auth cookie path as string.
func (m *ProtocolInfoResponse) AuthCookieFile() string {
    return m.Auth()["COOKIEFILE"]
}

// Get software version key value pairs in a map.
func (m *ProtocolInfoResponse) Version() map[string]string {
    text := __find_prefix_mrl(m.Buffer.MidReplyLines, "VERSION")
    return __make_variable_map(text[len("VERSION"):])
}

// Perform PROTOCOLINFO command request. Returns ProtocolInfoResponse instance
// reflecting command result.
func (c *Controller) ProtocolInfo() (*ProtocolInfoResponse, error) {
    request := NewRequest(COMMAND_PROTOCOLINFO)
    response := &ProtocolInfoResponse{}
    return response, c.Request(request, response)
}

// The GetConfResponse type is returned by the GetConf command method.
type GetConfResponse struct { *BaseControlResponse }

// Get singleton value from response and return it (short hand for single reqs.)
func (m *GetConfResponse) Value() string {
    //FIXME: Maybe just return first value?
    return strings.SplitN(m.Buffer.EndReplyLine.StatusText(), "=", 2)[1]
}

// Get specific value from configuration ``key'' from request.
func (m *GetConfResponse) ValueOf(key string) string {
    for _, i := range m.Buffer.MidReplyLines {
        parts := strings.SplitN(i.Text(), "=", 2)
        if parts[0] == key { return parts[1] }
    }

    return ""
}

// Get all values from configuration as a map.
func (m *GetConfResponse) ValueAll() map[string]string {
    results := make(map[string]string)
    // Collect results from MidReplyLines
    for _, i := range m.Buffer.MidReplyLines {
        parts := strings.SplitN(i.Text(), "=", 2)
        results[parts[0]] = parts[1]
    }
    // Append result from EndReplyLine
    parts := strings.SplitN(m.Buffer.EndReplyLine.StatusText(), "=", 2)
    results[parts[0]] = parts[1]
    return results
}

// Perform GETCONF command request. Returns GetConfResponse instance reflecting
// command result.
func (c *Controller) GetConf(keys []string) (*GetConfResponse, error) {
    request := NewRequest(COMMAND_GETCONF + " " + strings.Join(keys, " "))
    response := &GetConfResponse{}
    return response, c.Request(request, response)
}

// The SetConfResponse type is returned by the SetConf command method.
type SetConfResponse struct { *BaseControlResponse }

// Perform SETCONF command request. Returns SetConfResponse instance reflecting
// command result.
func (c *Controller) SetConf(opts map[string][]string) (*SetConfResponse, error) {
    kvs := make([]string, 0)
    for k, v := range opts {
        for _, j := range v {
            kvs = append(kvs, k + "=" + j)
        }
    }
    request := NewRequest(COMMAND_SETCONF + " " + strings.Join(kvs, " "))
    response := &SetConfResponse{}
    return response, c.Request(request, response)
}


// The ResetConfResponse type is returned by the ResetConf command method.
type ResetConfResponse struct { *BaseControlResponse }

// Perform RESETCONF command request. Returns ResetConfResponse instance
// reflecting command result.
func (c *Controller) ResetConf(opts map[string][]string) (*ResetConfResponse, error) {
    kvs := make([]string, 0)
    for k, v := range opts {
        for _, j := range v {
            kvs = append(kvs, k + "=" + j)
        }
    }
    request := NewRequest(COMMAND_RESETCONF + " " + strings.Join(kvs, " "))
    response := &ResetConfResponse{}
    return response, c.Request(request, response)
}

/*TODO: Implement multiline request support ....
type LoadConfResponse struct { *BaseControlResponse }
func (c *Controller) LoadConf() (*LoadConfResponse, error) {
    request := NewRequest(COMMAND_LOADCONF)
    response := &LoadConfResponse{}
    return response, c.Request(request, response)
}
*/

// The SaveConfResponse type is returned by the SaveConf command method.
type SaveConfResponse struct { *BaseControlResponse }

// Perform SAVECONF command request. Returns ResetConfResponse instance
// reflecting command result.
func (c *Controller) SaveConf() (*SaveConfResponse, error) {
    request := NewRequest(COMMAND_SAVECONF)
    response := &SaveConfResponse{}
    return response, c.Request(request, response)
}

// The SetEventsResponse type is returned by the SetEvents command method.
type SetEventsResponse struct { *BaseControlResponse }

// Perform SETEVENTS command request. Returns SetEventsResponse instance
// reflecting command result.
func (c *Controller) SetEvents(events []string) (*SetEventsResponse, error) {
    request := NewRequest(COMMAND_SETEVENTS + " " + strings.Join(events, " "))
    response := &SetEventsResponse{}
    return response, c.Request(request, response)
}

type Signal string

// Constants to use along with Signal command method.
const (
           SIGNAL_RELOAD = Signal("RELOAD")
         SIGNAL_SHUTDOWN = Signal("SHUTDOWN")
             SIGNAL_DUMP = Signal("DUMP")
            SIGNAL_DEBUG = Signal("DEBUG")
             SIGNAL_HALT = Signal("HALT")
              SIGNAL_HUP = Signal("HUP")
              SIGNAL_INT = Signal("INT")
             SIGNAL_USR1 = Signal("USR1")
             SIGNAL_USR2 = Signal("USR2")
             SIGNAL_TERM = Signal("TERM")
           SIGNAL_NEWNYM = Signal("NEWNYM")
    SIGNAL_CLEARDNSCACHE = Signal("CLEARDNSCACHE")
        SIGNAL_HEARTBEAT = Signal("HEARTBEAT")
)

// The SignalResponse type is returned by the Signal command method.
type SignalResponse struct { *BaseControlResponse }

// Perform SIGNAL command request. Returns SignalResponse instance reflecting
// command result.
func (c *Controller) Signal(signal Signal) (*SignalResponse, error) {
    request := NewRequest(COMMAND_SIGNAL + " " + string(signal))
    response := &SignalResponse{}
    return response, c.Request(request, response)
}

// The DropGuardsResponse type is returned by the DropGuards command method.
type DropGuardsResponse struct { *BaseControlResponse }

// Perform DROPGUARDS command request. Returns DropGuardsResponse instance
// reflecting command result.
func (c *Controller) DropGuards() (*DropGuardsResponse, error) {
    request := NewRequest(COMMAND_DROPGUARDS)
    response := &DropGuardsResponse{}
    return response, c.Request(request, response)
}

// The AddOnionResponse type is returned by the AddOnion command method.
type AddOnionResponse struct { *BaseControlResponse }

// Returns the ServiceID field of the created hidden service.
func (m *AddOnionResponse) ServiceId() string {
    v := __find_prefix_mrl(m.Buffer.MidReplyLines, "ServiceID=")
    return v[len("ServiceID"):]
}

// Returns the PrivateKey field of the created hidden service.
func (m *AddOnionResponse) PrivateKey() string {
    v := __find_prefix_mrl(m.Buffer.MidReplyLines, "PrivateKey=")
    return v[len("PrivateKey"):]
}

// Constants to use with the AddOnion command method.
const (
           ONION_KEY_TYPE_NEW = "NEW"
       ONION_KEY_TYPE_RSA1024 = "RSA1024"

          ONION_KEY_BLOB_BEST = "BEST"
       ONION_KEY_BLOB_RSA1024 = "RSA1024"

    ADD_ONION_FLAG_DISCARD_PK = "DiscardPK"
        ADD_ONION_FLAG_DETACH = "Detach"
)

// Perform ADD_ONION command request. Returns AddOnionResponse instance
// reflecting command result.
func (c *Controller) AddOnion(keyType string,
                              keyData string,
                              flags []string,
                              ports []string) (*AddOnionResponse, error) {
    reqline := COMMAND_ADD_ONION + " " + string(keyType) + ":" + keyData

    if len(flags) > 0 {
        reqline += " Flags=" + strings.Join(flags, ",")
    }

    for _, v := range ports {
        reqline += " Port=" + v
    }

    request := NewRequest(reqline)
    response := &AddOnionResponse{}
    return response, c.Request(request, response)
}

// The DelOnionResponse type is returned by the DelOnion command method.
type DelOnionResponse struct { *BaseControlResponse }

// Perform DEL_ONION command request. Returns DelOnionResponse instance
// reflecting command result.
func (c *Controller) DelOnion(serviceId string) (*DelOnionResponse, error) {
    if strings.HasSuffix(serviceId, ".onion") {
        serviceId = strings.TrimSuffix(serviceId, ".onion")
    }
    request := NewRequest(COMMAND_DEL_ONION + " " + serviceId)
    response := &DelOnionResponse{}
    return response, c.Request(request, response)
}

// Helpers ---------------------------------------------------------------------
// TODO: Move into parser, or buffer.go
func __find_prefix_mrl(data []MidReplyLine, prefix string) string {
    for _, v := range data {
        if strings.HasPrefix(v.Text(), prefix) { return v.Text() }
    }
    return ""
}

func __find_prefix_drl(data []DataReplyLine, prefix string) string {
    for _, v := range data {
        if strings.HasPrefix(v.Text(), prefix) {
            return strings.Join(v[1:], "\n")
        }
    }
    return ""
}

func __make_variable_map(data string) map[string]string {
    results := make(map[string]string)
    r := regexp.MustCompile(`([/\w]+)=(([,\w]+)|("(?:\\.|[^"\\])*)")`)

    for _, j := range r.FindAllStringSubmatch(data, -1) {
        results[j[1]] = strings.TrimPrefix(strings.TrimSuffix(j[2], "\""), "\"")
    }

    return results
}

