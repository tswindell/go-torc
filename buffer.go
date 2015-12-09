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
    "strconv"
    "strings"
)


type MidReplyLine   string
type DataReplyLine []string
type EndReplyLine   string

// The LineBuffer type is used when building request messages.
type LineBuffer []string

// The ResponseBuffer type is used by the parser to build incoming messages.
type ResponseBuffer struct {
    MidReplyLines  []MidReplyLine
    DataReplyLines []DataReplyLine
    EndReplyLine    EndReplyLine
}

// Returns the status integer in a MidReplyLine.
func (l MidReplyLine) Status() int {
    i, e := strconv.Atoi(l._parts()[0])
    if e != nil { return -1 }
    return i
}

// Returns the text segment of a MidReplyLine.
func (l MidReplyLine) Text() string {
    return l._parts()[1]
}

func (l MidReplyLine) _parts() []string {
    return strings.SplitN(string(l), "-", 2)
}

// Returns the status integer in a DataReplyLine
func (l DataReplyLine) Status() int {
    i, e := strconv.Atoi(l._parts()[0])
    if e != nil { return -1 }
    return i
}

// Returns the text segment of a DataReplyLine
func (l DataReplyLine) Text() string {
    return l._parts()[1]
}

func (l DataReplyLine) _parts() []string {
    return strings.SplitN(string(l[0]), "+", 2)
}

// Returns Status code of a response.
func (l EndReplyLine) Status() int {
    i, e := strconv.Atoi(l._status_parts()[0])
    if e != nil { return -1 }
    return i
}

// Returns the Status code text message of a response.
func (l EndReplyLine) StatusText() string {
    return l._status_parts()[1]
}

func (l EndReplyLine) _status_parts() []string {
    return strings.SplitN(string(l), " ", 2)
}

// Makes sure that the message is formatted correctly, returns []byte for
// transmission over control channel.
func (b LineBuffer) Normalize() []byte {
    results := make([]byte, 0)
    for _, v := range b {
        // Normalize newlines
        v = strings.Replace(v, "\r\n", "\n", -1)
        v = strings.Replace(v, "\n", "\r\n", -1)
        // Append newline suffix to line.
        if !strings.HasSuffix(v, "\r\n") { v += "\r\n" }
        // Add to byte arrray.
        results = append(results, []byte(v)...)
    }
    return results
}

