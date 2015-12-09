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
    "bufio"
    "io"
    "strconv"
    "strings"
)

type Parser struct {
    reader *bufio.Reader

    ch chan ResponseBuffer

    // Parser state.
    buffer *ResponseBuffer
    dataReplyLine DataReplyLine

    bufferRaw []string

    isReady bool
    isMultiLine bool
}

func NewParser(r io.Reader, out chan ResponseBuffer) *Parser {
    p := new(Parser)
    p.reader = bufio.NewReader(r)
    p.ch = out
    p.Reset()
    return p
}

// Perform parser state reset.
func (p *Parser) Reset() {
    p.buffer = new(ResponseBuffer)
    p.buffer.MidReplyLines = make([]MidReplyLine, 0)
    p.buffer.DataReplyLines = make([]DataReplyLine, 0)

    p.bufferRaw = make([]string, 0)

    p.isReady = true
    p.isMultiLine = false
}

// Perform post to channel.
func (p *Parser) post() {
    LogComms(">>", p.bufferRaw)
    p.ch<- *p.buffer
    p.Reset()
}

// Run parser loop on reader.
func (p *Parser) Run() {
    for {
        ln, e := p.reader.ReadString('\n')
        if e == io.EOF {
            break // If we've got EOF, then quit reader.
        } else if e != nil {
            // For debugging, if we get here then we need to add more conds.
            LogWarn("Error reading on socket: %v", e)
            continue
        }

        // Check for and remove line ending from stanza.
        // TODO: Should probably buffer here to create full line.
        if !strings.HasSuffix(ln, "\r\n") {
            LogWarn("Protocol error, no line ending!")
            continue
        }

        // Trim incoming data and insert it into message data buffer.
        ln = strings.TrimSuffix(ln, "\r\n")
        p.bufferRaw = append(p.bufferRaw, ln)

        //   Check to see if we're waiting for a new message. If we are, then we
        // should expect 3 different possible formats:
        if p.isReady {
            switch {
            case isEndReplyLine(ln):
                p.buffer.EndReplyLine = EndReplyLine(ln)
                p.post()
                continue

            case isMidReplyLine(ln):
                p.buffer.MidReplyLines = append(p.buffer.MidReplyLines, MidReplyLine(ln))
                continue

            case isDataReplyLine(ln):
                p.dataReplyLine = DataReplyLine{ln}
                p.isReady = false
                p.isMultiLine = true
                continue

            default:
                LogError("Failed to parse initial message line!")
                p.Reset()
                continue
            }
        }

        if p.isMultiLine {
            if !isEndOfData(ln) {
                p.dataReplyLine = append(p.dataReplyLine, ln)
                continue
            }

            p.buffer.DataReplyLines = append(p.buffer.DataReplyLines, p.dataReplyLine)

            p.isReady = true
            p.isMultiLine = false
            continue
        }
    }
}

// Stateless parsing helpers ---------------------------------------------------
func isReplyLine(ln string) bool {
    if len(ln) < 5 { return false }

    if _, e := strconv.Atoi(ln[:3]); e != nil {
        return false
    }

    return true
}

func isEndReplyLine(ln string) bool {
    if !isReplyLine(ln) || ln[3] != ' ' { return false }
    return true
}

func isMidReplyLine(ln string) bool {
    if !isReplyLine(ln) || ln[3] != '-' { return false }
    return true
}

func isDataReplyLine(ln string) bool {
    if !isReplyLine(ln) || ln[3] != '+' { return false }
    return true
}

func isEndOfData(ln string) bool {
    if len(ln) != 1 || ln[0] != '.' { return false }
    return true
}

