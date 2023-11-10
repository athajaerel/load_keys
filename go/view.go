package main

import (
	"os"
	"fmt"
	"time"
)

type _loglevel int32

const (
	ERROR = _loglevel(0)
	WARN  = _loglevel(1)
	INFO  = _loglevel(2)
	DEBUG = _loglevel(3)
)

var prefixes = [4]string {"[OwO]", "[o.O]", "[o.o]", "[-.-]"}

const timeFmt = "2006-01-02 15:04:05.000000 UTC "

type View struct {
	loglevel _loglevel
	HasTime bool
	HasPrefix bool
}

func (re View) log(i _loglevel, s string) {
	if re.loglevel >= i {
		if re.HasTime {
			fmt.Print(time.Now().UTC().Format(timeFmt))
		}
		if re.HasPrefix {
			fmt.Print(prefixes[i])
			fmt.Print(" ")
		}
		fmt.Println(s)
		os.Stdout.Sync()
	}
}
