package main

import (
	"fmt"

	"github.com/ogsts/gortsplib"
	"github.com/ogsts/gortsplib/pkg/base"
	"github.com/ogsts/gortsplib/pkg/headers"
)

// This example shows how to
// 1. connect to a RTSP server
// 2. get tracks published on a path
// 3. read only selected tracks

func main() {
	u, err := base.ParseURL("rtsp://myserver/mypath")
	if err != nil {
		panic(err)
	}

	conn, err := gortsplib.Dial(u.Scheme, u.Host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.Options(u)
	if err != nil {
		panic(err)
	}

	tracks, baseURL, _, err := conn.Describe(u)
	if err != nil {
		panic(err)
	}

	// start reading only video tracks, skipping audio or application tracks
	for _, t := range tracks {
		if t.Media.MediaName.Media == "video" {
			_, err := conn.Setup(headers.TransportModePlay, baseURL, t, 0, 0)
			if err != nil {
				panic(err)
			}
		}
	}

	// play setupped tracks
	_, err = conn.Play(nil)
	if err != nil {
		panic(err)
	}

	// read RTP packets
	err = conn.ReadFrames(func(trackID int, streamType gortsplib.StreamType, payload []byte) {
		fmt.Printf("frame from track %d, type %v, size %d\n", trackID, streamType, len(payload))
	})
	panic(err)
}
