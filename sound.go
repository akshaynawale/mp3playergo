package main

import (
	"fmt"
	"github.com/golang/glog"
	"os/exec"
)

func SetVol(percent int) {
	var vol int
	if percent != 0 {
		vol = 60 + ((percent / 10) * 4)
	} else {
		vol = 60
	}
	cmd := exec.Command("amixer", "sset", "\"PCM\"", fmt.Sprintf("%d%%", vol))
	//fmt.Printf("amixer sset \"PCM\" %d%", vol)
	err := cmd.Run()
	if err != nil {
		glog.Infof("%s\n", err)
	}
}
