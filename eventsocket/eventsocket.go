// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// FreeSWITCH Event Socket library for the Go programming language.
//
// eventsocket supports both inbound and outbound event socket connections,
// acting either as a client connecting to FreeSWITCH or as a server accepting
// connections from FreeSWITCH to control calls.
//
// Reference:
// http://wiki.freeswitch.org/wiki/Event_Socket
// http://wiki.freeswitch.org/wiki/Event_Socket_Outbound
//
// WORK IN PROGRESS, USE AT YOUR OWN RISK.
package eventsocket

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"net/url"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const bufferSize = 1024 << 6 // For the socket reader
const eventsBuffer = 16      // For the events channel (memory eater!)
const timeoutPeriod = 60 * time.Second

var errMissingAuthRequest = errors.New("Missing auth request")
var errInvalidPassword = errors.New("Invalid password")
var errInvalidCommand = errors.New("Invalid command contains \\r or \\n")
var errTimeout = errors.New("Timeout")

// Connection is the event socket connection handler.
type Connection struct {
	conn          net.Conn
	reader        *bufio.Reader
	textreader    *textproto.Reader
	errEv, errReq chan error
	cmd, api, evt chan []string
}

// newConnection allocates a new Connection and initialize its buffers.
func newConnection(c net.Conn) *Connection {
	h := Connection{
		conn:   c,
		reader: bufio.NewReaderSize(c, bufferSize),
		errEv:  make(chan error, 1),
		errReq: make(chan error, 1),
		cmd:    make(chan []string),
		api:    make(chan []string),
		evt:    make(chan []string, eventsBuffer),
	}
	h.textreader = textproto.NewReader(h.reader)
	return &h
}

// HandleFunc is the function called on new incoming connections.
type HandleFunc func(*Connection)

// ListenAndServe listens for incoming connections from FreeSWITCH and calls
// HandleFunc in a new goroutine for each client.
//
// Example:
//
//	func main() {
//		eventsocket.ListenAndServe(":9090", handler)
//	}
//
//	func handler(c *eventsocket.Connection) {
//		ev, err := c.Send("connect") // must always start with this
//		ev.PrettyPrint()             // print event to the console
//		...
//		c.Send("myevents")
//		for {
//			ev, err = c.ReadEvent()
//			...
//		}
//	}
func ListenAndServe(addr string, fn HandleFunc) error {
	srv, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		c, err := srv.Accept()
		if err != nil {
			return err
		}
		h := newConnection(c)
		go h.readLoop()
		go fn(h)
	}
}

// Dial attemps to connect to FreeSWITCH and authenticate.
//
// Example:
//
//	c, _ := eventsocket.Dial("localhost:8021", "ClueCon")
//	ev, _ := c.Send("events plain ALL") // or events json ALL
//	for {
//		ev, _ = c.ReadEvent()
//		ev.PrettyPrint()
//		...
//	}
func Dial(addr, passwd string) (*Connection, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	h := newConnection(c)
	m, err := h.textreader.ReadMIMEHeader()
	if err != nil {
		c.Close()
		return nil, err
	}
	if m.Get("Content-Type") != "auth/request" {
		c.Close()
		return nil, errMissingAuthRequest
	}
	fmt.Fprintf(c, "auth %s\r\n\r\n", passwd)
	m, err = h.textreader.ReadMIMEHeader()
	if err != nil {
		c.Close()
		return nil, err
	}
	if m.Get("Reply-Text") != "+OK accepted" {
		c.Close()
		return nil, errInvalidPassword
	}
	go h.readLoop()
	return h, err
}

// readLoop calls readOne until a fatal error occurs, then close the socket.
func (h *Connection) readLoop() {
	for h.readOne() {
	}
	h.Close()
}

// readOne reads a single event and send over the appropriate channel.
// It separates incoming events from api and command responses.
func (h *Connection) readOne() bool {
	var (
		err    error
		length int
		hdr    textproto.MIMEHeader
	)

	resp := make([]string, FsEventMapSize)
	hdr, err = h.textreader.ReadMIMEHeader()
	if err != nil {
		h.errEv <- err
		return false
	}

	if v := hdr.Get("Content-Length"); v != "" {
		length, err = strconv.Atoi(v)
		if err == nil {
			b := make([]byte, length)
			if _, err = io.ReadFull(h.reader, b); err == nil {
				resp[EventMsgBody] = string(b)
			}
		}
	}

	switch hdr.Get("Content-Type") {
	case "command/reply":
		if err != nil {
			h.errReq <- err
			return false
		}
		reply := hdr.Get("Reply-Text")
		if len(reply) > 1 && reply[:2] == "-E" {
			h.errReq <- errors.New(reply[5:])
			return true
		}
		if reply[0] == '%' {
			copyHeaders(&hdr, resp, true)
		} else {
			copyHeaders(&hdr, resp, false)
		}
		h.cmd <- resp
	case "api/response":
		if err != nil {
			h.errReq <- err
			return false
		}
		if len(resp[EventMsgBody]) > 1 && string(resp[EventMsgBody][:2]) == "-E" {
			h.errReq <- errors.New(string(resp[EventMsgBody])[5:])
			return true
		}
		copyHeaders(&hdr, resp, false)
		h.api <- resp
	case "text/event-plain":
		if err != nil {
			h.errEv <- err
			return false
		}
		reader := bufio.NewReader(bytes.NewReader([]byte(resp[EventMsgBody])))
		resp[EventMsgBody] = ""
		textreader := textproto.NewReader(reader)
		hdr, err = textreader.ReadMIMEHeader()
		if err != nil {
			h.errEv <- err
			return false
		}
		if v := hdr.Get("Content-Length"); v != "" {
			length, err := strconv.Atoi(v)
			if err != nil {
				h.errEv <- err
				return false
			}
			b := make([]byte, length)
			if _, err = io.ReadFull(reader, b); err != nil {
				h.errEv <- err
				return false
			}
			resp[EventMsgBody] = string(b)
		}
		copyHeaders(&hdr, resp, true)
		h.evt <- resp
	// case "text/event-json":
	// 	h.errEv <- errors.New("text/event-json: unsupported ")
	// 	if err != nil {
	// 		h.errEv <- err
	// 		return false
	// 	}
	// 	tmp := make(EventHeader)
	// 	err := json.Unmarshal([]byte(resp.Body), &tmp)
	// 	if err != nil {
	// 		h.errEv <- err
	// 		return false
	// 	}
	// 	// capitalize header keys for consistency.
	// 	for k, v := range tmp {
	// 		resp.Header[capitalize(k)] = v
	// 	}
	// 	if v, _ := resp.Header["_body"]; v != nil {
	// 		resp.Body = v.(string)
	// 		delete(resp.Header, "_body")
	// 	} else {
	// 		resp.Body = ""
	// 	}
	// 	h.evt <- resp
	case "text/disconnect-notice":
		if err != nil {
			h.errEv <- err
			return false
		}
		copyHeaders(&hdr, resp, false)
		h.evt <- resp
	default:
		h.errEv <- errors.New("unsupported event")
		log.Fatal("Unsupported event:", hdr)
	}
	return true
}

// RemoteAddr returns the remote addr of the connection.
func (h *Connection) RemoteAddr() net.Addr {
	return h.conn.RemoteAddr()
}

// Close terminates the connection.
func (h *Connection) Close() {
	h.conn.Close()
}

// ReadEvent reads and returns events from the server. It supports both plain
// or json, but *not* XML.
//
// When subscribing to events (e.g. `Send("events json ALL")`) it makes no
// difference to use plain or json. ReadEvent will parse them and return
// all headers and the body (if any) in an Event struct.
func (h *Connection) ReadEvent() ([]string, error) {
	var (
		ev  []string
		err error
	)
	select {
	case err = <-h.errEv:
		return nil, err
	case ev = <-h.evt:
		return ev, nil
	}
}

// copyHeaders copies all keys and values from the MIMEHeader to Event.Header,
// normalizing header keys to their capitalized version and values by
// unescaping them when decode is set to true.
//
// It's used after parsing plain text event headers, but not JSON.
func copyHeaders(src *textproto.MIMEHeader, dst []string, decode bool) {
	var err error
	var customHeaderValues string
	clen := len(customHeaderPrefix)
	for k, v := range *src {
		k = capitalize(k)
		if decode && MapKeyIndex[k] != 0 {
			dst[MapKeyIndex[k]], err = url.QueryUnescape(v[0])
			if err != nil {
				dst[MapKeyIndex[k]] = v[0]
			}
		} else if MapKeyIndex[k] != 0 {
			dst[MapKeyIndex[k]] = v[0]
		} else if decode && len(k) >= clen && k[0:clen] == customHeaderPrefix {
			var headerValue string
			headerValue, err = url.QueryUnescape(v[0])
			if err != nil {
				headerValue = v[0]
			}
			customHeaderValues = customHeaderValues + k + ":" + headerValue + "|"
		} else if len(k) >= clen && k[0:clen] == customHeaderPrefix {
			customHeaderValues = customHeaderValues + k + ":" + v[0] + "|"
		}
	}

	if customHeaderValues != "" {
		dst[CustomHeaders] = customHeaderValues
	}
}

// capitalize capitalizes strings in a very particular manner.
// Headers such as Job-UUID become Job-Uuid and so on. Headers starting with
// Variable_ only replace ^v with V, and headers staring with _ are ignored.
func capitalize(s string) string {
	if s[0] == '_' {
		return s
	}
	ns := bytes.ToLower([]byte(s))
	if len(s) > 9 && s[1:9] == "ariable_" {
		ns[0] = 'V'
		return string(ns)
	}
	toUpper := true
	for n, c := range ns {
		if toUpper {
			if 'a' <= c && c <= 'z' {
				c -= 'a' - 'A'
			}
			ns[n] = c
			toUpper = false
		} else if c == '-' || c == '_' {
			toUpper = true
		}
	}
	return string(ns)
}

// Send sends a single command to the server and returns a response Event.
//
// See http://wiki.freeswitch.org/wiki/Event_Socket#Command_Documentation for
// details.
func (h *Connection) Send(command string) ([]string, error) {
	// Sanity check to avoid breaking the parser
	//if strings.IndexAny(command, "\r\n") > 0 {
	//	return nil, errInvalidCommand
	//}
	fmt.Fprintf(h.conn, "%s\r\n\r\n", command)
	var (
		ev  []string
		err error
	)
	select {
	case err = <-h.errReq:
		return nil, err
	case ev = <-h.cmd:
		fmt.Println("Send: rcvd event from cmd channel")
		return ev, nil
	case ev = <-h.api:
		fmt.Println("Send: rcvd event from api channel")
		return ev, nil
	case <-time.After(timeoutPeriod):
		buf := make([]byte, 1<<20)
		runtime.Stack(buf, true)
		fmt.Println("", "Send - runtime stack: ", buf)
		return nil, errTimeout
	}
}

// MSG is the container used by SendMsg to store messages sent to FreeSWITCH.
// It's supposed to be populated with directives supported by the sendmsg
// command only, like "call-command: execute".
//
// See http://wiki.freeswitch.org/wiki/Event_Socket#sendmsg for details.
type MSG map[string]string

// SendMsg sends messages to FreeSWITCH and returns a response Event.
//
// Examples:
//
//	SendMsg(MSG{
//		"call-command": "hangup",
//		"hangup-cause": "we're done!",
//	}, "", "")
//
//	SendMsg(MSG{
//		"call-command":     "execute",
//		"execute-app-name": "playback",
//		"execute-app-arg":  "/tmp/test.wav",
//	}, "", "")
//
// Keys with empty values are ignored; uuid and appData are optional.
// If appData is set, a "content-length" header is expected (lower case!).
//
// See http://wiki.freeswitch.org/wiki/Event_Socket#sendmsg for details.
func (h *Connection) SendMsg(m MSG, uuid, appData string) ([]string, error) {
	b := bytes.NewBufferString("sendmsg")
	if uuid != "" {
		// Make sure there's no \r or \n in the UUID.
		if strings.IndexAny(uuid, "\r\n") > 0 {
			return nil, errInvalidCommand
		}
		b.WriteString(" " + uuid)
	}
	b.WriteString("\n")
	for k, v := range m {
		// Make sure there's no \r or \n in the key, and value.
		if strings.IndexAny(k, "\r\n") > 0 {
			return nil, errInvalidCommand
		}
		if v != "" {
			if strings.IndexAny(v, "\r\n") > 0 {
				return nil, errInvalidCommand
			}
			b.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
	}
	b.WriteString("\n")
	if m["content-length"] != "" && appData != "" {
		b.WriteString(appData)
	}
	if _, err := b.WriteTo(h.conn); err != nil {
		return nil, err
	}
	var (
		ev  []string
		err error
	)
	select {
	case err = <-h.errReq:
		return nil, err
	case ev = <-h.cmd:
		fmt.Println("SendMsg: rcvd event from cmd channel")
		return ev, nil
	case ev = <-h.api:
		fmt.Println("SendMsg: rcvd event from api channel")
		return ev, nil
	case <-time.After(timeoutPeriod):
		buf := make([]byte, 1<<20)
		runtime.Stack(buf, true)
		fmt.Println("", "SendMsg - runtime stack: ", buf)
		return nil, errTimeout
	}
}

// Execute is a shortcut to SendMsg with call-command: execute without UUID,
// suitable for use on outbound event socket connections (acting as server).
//
// Example:
//
//	Execute("playback", "/tmp/test.wav", false)
//
// See http://wiki.freeswitch.org/wiki/Event_Socket#execute for details.
func (h *Connection) Execute(appName, appArg string, lock bool) ([]string, error) {
	var evlock string
	if lock {
		// Could be strconv.FormatBool(lock), but we don't want to
		// send event-lock when it's set to false.
		evlock = "true"
	}
	return h.SendMsg(MSG{
		"call-command":     "execute",
		"execute-app-name": appName,
		"execute-app-arg":  appArg,
		"event-lock":       evlock,
	}, "", "")
}

// ExecuteUUID is similar to Execute, but takes a UUID and no lock. Suitable
// for use on inbound event socket connections (acting as client).
func (h *Connection) ExecuteUUID(uuid, appName, appArg, appUUID string) ([]string, error) {
	return h.SendMsg(MSG{
		"call-command":     "execute",
		"execute-app-name": appName,
		"execute-app-arg":  appArg,
		"event-uuid":       appUUID,
	}, uuid, "")
}

// EventHeader represents events as a pair of key:value.
type EventHeader map[string]interface{}

// Event represents a FreeSWITCH event.
type Event struct {
	Header EventHeader // Event headers, key:val
	Body   string      // Raw body, available in some events
}

func (r *Event) String() string {
	if r.Body == "" {
		return fmt.Sprintf("%s", r.Header)
	} else {
		return fmt.Sprintf("%s body=%s", r.Header, r.Body)
	}
}

// Get returns an Event value, or "" if the key doesn't exist.
func (r *Event) Get(key string) string {
	val, ok := r.Header[key]
	if !ok || val == nil {
		return ""
	}
	if s, ok := val.([]string); ok {
		return strings.Join(s, ", ")
	}
	return val.(string)
}

// GetInt returns an Event value converted to int, or an error if conversion
// is not possible.
func (r *Event) GetInt(key string) (int, error) {
	n, err := strconv.Atoi(r.Header[key].(string))
	if err != nil {
		return 0, err
	}
	return n, nil
}

// PrettyPrint prints Event headers and body to the standard output.
func (r *Event) PrettyPrint() {
	var keys []string
	for k := range r.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s: %#v\n", k, r.Header[k])
	}
	if r.Body != "" {
		fmt.Printf("BODY: %#v\n", r.Body)
	}
}
