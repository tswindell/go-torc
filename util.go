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
    "log"
    "runtime"
    "strconv"
    "strings"
)


func NotImplemented() error { return fmt.Errorf("Not implemented yet!") }

func elidePath(source string, maxlen int) string {
    if len(source) > maxlen {
        r := source[len(source)-maxlen:len(source)]
        return strings.SplitN(r, "/", 2)[1]
    }
    return source
}

func _log(t, m string, v ...interface{}) {
    pc, _, ln, _ := runtime.Caller(2)
    fn := runtime.FuncForPC(pc).Name()
    log.Printf(fn + ":" + strconv.Itoa(ln) + " -- " + t + " -- " + m + "\n", v...)
}

func LogInfo(m string, v ...interface{}) { _log("INFO", m, v...) }
func LogWarn(m string, v ...interface{}) { _log("WARN", m, v...) }
func LogError(m string, v ...interface{}) { _log("ERROR", m, v...) }

var COMMS_LOGGING = false
func LogComms(dir string, data LineBuffer) {
    if !COMMS_LOGGING { return }

    _log("COMMS", "\n " + dir + " " + strings.Join(strings.Split(string(data.Normalize()), "\r\n")[0:len(data)], "\n " + dir + " "))
}

