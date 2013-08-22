RingBuffer for Go
- Large sections implemented in C using CGO

[![Build Status](https://drone.io/github.com/Vizidrix/ringbuffer/status.png)](https://drone.io/github.com/Vizidrix/ringbuffer/latest)

## Introduction ##

----

> I've been striding across this buffer all day and still haven't found the end, how much memory does this thing have?

  - Gotta have [Go]: http://golang.org

----

This library is very largely inspired by the contributions of the LMAX dev team which built Disruptor

## Features ##

Batchable allocation of buffer entries
Transaction support via batch claim, publish and release
Zero copy buffer population
Lock-free for single producer, multi consumer scenario
* Future support for multi producer fan-in planned


A few unit tests to try and prove the functionality of the buffer.

## Getting Started ##

1\. Add the correct import for your project.

```go
import (
	"github.com/vizidrix/ringbuffer"
)
```

2\. Start using. (** API is in active development and is likely to change)

```
ringSize := L0 // See translation to size in ringbuffer.go
chunkCount := 2 // how many 32byte chunks for each buffer slot?

buffer := NewRingBuffer(ringSize, chunkCount)

// Can read details of what was partitioned by:
info := buffer.GetInfo()

// Get a chunk of data from somewhere
data := make([]byte, 64)
... populate the data slice

// Determine how many entries you intend to publish
batchSize := 1

// Ask for a zero-copy buffer to write into
batch := buffer.Claim(batchSize)

// This example creates a buffer but in many scenarios you'll
// be pulling directly from a stream, file, web request, etc
batch.Entries[0].CopyFrom(data)

// Done filling up the buffer
token := batch.Publish()

// If you care to make sure your publish worked
select {
  case <-token.Published: {
    // Handle success
  }
  case <-token.Failed: {
    // Handle failure
  }
  case <-time.After(10 * time.Milliseconds): {
    // Handle timeout
  }
}
```

> Read side docs comming soon...


(Real docs pending)

# Goals #
- Move data into a target system very quickly with as little blocking or delay possible
- Sufficient semantics to support a reading database implementation with transaction support and possibly ACID


----

Version
----
0.1.0 ish

Tech
----

* [Go] - Golang.org
* [RINGBUFFER] - Ring Buffer (Disruptor) for Go

License
----

https://github.com/Vizidrix/gocqrs/blob/master/LICENSE

----
## Edited
* 21-Aug-2013	initial documentation

----
## Credits
* Vizidrix <https://github.com/organizations/Vizidrix>
* Perry Birch <https://github.com/PerryBirch>
