package base

import (
    "time"
)

type EventHandler struct {
    // Opened fires when a new connection has opened.
    // The info parameter has information about the connection such as
    // it's local and remote address.
    // Use the out return value to write data to the connection.
    // The opts return value is used to set connection options.
    OnConnected func(id int, info Info) (out []byte, opts Options, action Action)

    // Closed fires when a connection has closed.
    // The err parameter is the last known connection error.
    Closed func(id int, err error) (action Action)

    // Tick fires immediately after the server starts and will fire again
    // following the duration specified by the delay return value.
    OnTimeout func() (delay time.Duration, action Action)
}
