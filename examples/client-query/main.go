package main

import (
	"fmt"

	"github.com/ogsts/gortsplib"
	"github.com/ogsts/gortsplib/pkg/base"
)

// This example shows how to
// 1. connect to a RTSP server
// 2. get and print informations about tracks published on a path.

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

	tracks, _, _, err := conn.Describe(u)
	if err != nil {
		panic(err)
	}

	fmt.Println("available tracks: %v\n", tracks)
}
