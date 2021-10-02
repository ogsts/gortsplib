
# gortsplib

[![Test](https://github.com/ogsts/gortsplib/workflows/test/badge.svg)](https://github.com/ogsts/gortsplib/actions?query=workflow:test)
[![Lint](https://github.com/ogsts/gortsplib/workflows/lint/badge.svg)](https://github.com/ogsts/gortsplib/actions?query=workflow:lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/ogsts/gortsplib)](https://goreportcard.com/report/github.com/ogsts/gortsplib)
[![CodeCov](https://codecov.io/gh/ogsts/gortsplib/branch/main/graph/badge.svg)](https://codecov.io/gh/ogsts/gortsplib/branch/main)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/ogsts/gortsplib)](https://pkg.go.dev/github.com/ogsts/gortsplib#pkg-index)

RTSP 1.0 client and server library for the Go programming language, written for [rtsp-simple-server](https://github.com/ogsts/rtsp-simple-server).

Go &ge; 1.15 is required.

Features:

* Client
  * Query servers about available streams
  * Read
    * Read streams from servers with UDP, UDP-multicast, TCP or TLS
    * Switch protocol automatically (switch to TCP in case of server error or UDP timeout)
    * Read only selected tracks of a stream
    * Pause or seek without disconnecting from the server
    * Generate RTCP receiver reports automatically
  * Publish
    * Publish streams to servers with UDP, TCP or TLS
    * Switch protocol automatically (switch to TCP in case of server error)
    * Pause without disconnecting from the server
    * Generate RTCP sender reports automatically
* Server
  * Handle requests from clients
  * Sessions and connections are independent
  * Read streams from clients with UDP, TCP or TLS
  * Write streams to clients with UDP, UDP-multicast, TCP or TLS
  * Provide SSRC, RTP-Info to clients automatically
  * Generate RTCP receiver reports automatically
* Utilities
  * Encode and decode RTSP primitives, RTP/H264, RTP/AAC, SDP

## Table of contents

* [Examples](#examples)
* [API Documentation](#api-documentation)
* [Links](#links)

## Examples

* [client-query](examples/client-query/main.go)
* [client-read](examples/client-read/main.go)
* [client-read-partial](examples/client-read-partial/main.go)
* [client-read-options](examples/client-read-options/main.go)
* [client-read-pause](examples/client-read-pause/main.go)
* [client-read-h264](examples/client-read-h264/main.go)
* [client-read-save-to-disk](examples/client-read-save-to-disk/main.go)
* [client-publish](examples/client-publish/main.go)
* [client-publish-options](examples/client-publish-options/main.go)
* [client-publish-pause](examples/client-publish-pause/main.go)
* [server](examples/server/main.go)
* [server-tls](examples/server-tls/main.go)

## API Documentation

https://pkg.go.dev/github.com/ogsts/gortsplib#pkg-index

## Links

Related projects

* https://github.com/ogsts/rtsp-simple-server
* https://github.com/pion/sdp (SDP library used internally)
* https://github.com/pion/rtcp (RTCP library used internally)
* https://github.com/pion/rtp (RTP library used internally)

IETF Standards

* RTSP 1.0 https://tools.ietf.org/html/rfc2326
* RTSP 2.0 https://tools.ietf.org/html/rfc7826
* HTTP 1.1 https://tools.ietf.org/html/rfc2616

Conventions

* https://github.com/golang-standards/project-layout
