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

package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/tswindell/go-torc"
)

var (
    hostport = flag.String("hostport", "127.0.0.1:9051",
               "Remote host to connect to (default 127.0.0.1:9051)")
)

func NotImplemented() error { return fmt.Errorf("Not implemented yet!") }

// Application Entry-Point -----------------------------------------------------
func main() {
    flag.Parse()

    torc.LogInfo("Initializing control socket service handler.")
    ctrl := torc.NewController("tcp", *hostport)
    defer ctrl.Close()

    torc.LogInfo("Attempting to connect to Tor control socket.")
    if ctrl.Connect() != nil {
        os.Exit(1)
    }

    p, e := ctrl.ProtocolInfo()
    if e != nil || !p.IsSuccess() {
        torc.LogInfo("Failed to send protocol info request: %v", e)
        os.Exit(1)
    }
    torc.LogInfo("PROTOCOLINFO=%d", p.Protocol())
    torc.LogInfo("AUTH-METHODS=%v", p.AuthMethods())
    torc.LogInfo("AUTH-COOKIEFILE=%v", p.AuthCookieFile())
    torc.LogInfo("VERSIONS Tor=%v", p.Version()["Tor"])

    r, e := ctrl.GetInfo([]string{"version", "info/names", "events/names"})
    if e != nil || !r.IsSuccess() {
        torc.LogInfo("Failed to send get info request: %v", e)
        os.Exit(1)
    }

    torc.LogInfo("  Tor VERSION=%s", r.ValueOf("version"))
    torc.LogInfo("  info/names=%s", r.ValueOf("info/names"))
    torc.LogInfo("  events/names=%s", r.ValueOf("events/names"))

    c, e := ctrl.GetConf([]string{"SocksPort", "ControlPort", "CookieAuthentication"})
    if e != nil || !c.IsSuccess() {
        torc.LogInfo("Failed to send get configuration request: %v", e)
        os.Exit(1)
    }
    torc.LogInfo("%v", c.ValueAll())

    d, e := ctrl.SetConf(map[string][]string{
        "SocksPort": []string{"127.0.0.1:9050", "10.0.0.1:9050"},
    })
    if e != nil {
        torc.LogInfo("Failed to send set configuration request: %v", e)
        os.Exit(1)
    }
    torc.LogInfo("%d - %s", d.Status(), d.StatusText())

    o, e := ctrl.AddOnion(torc.ONION_KEY_TYPE_NEW,
                          torc.ONION_KEY_BLOB_BEST,

                          []string{
                              torc.ADD_ONION_FLAG_DETACH,
                          },

                          []string{
                              "80,127.0.0.1:9600",
                          })
    if e != nil {
        torc.LogError("Failed to send add onion request: %v", e)
        os.Exit(1)
    }
    torc.LogInfo("  ServiceID: %s", o.ServiceId())
    torc.LogInfo(" ServiceKey: %s", o.PrivateKey())

    os.Exit(0)
}

