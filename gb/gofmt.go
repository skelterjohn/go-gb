package main

import (
	"exec"
	"os"
	"fmt"
)

func RunGoFMT(dir, file string) (err os.Error) {
	margs := []string{"gofmt", "-w", file}

	if Verbose {
		fmt.Printf("%v\n", margs)
	}

	p, err := exec.Run(GoFMTCMD, margs, os.Envs, dir, exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		return
	}
	if p != nil {
		p.Wait(0)
	}
	return
	return
}
