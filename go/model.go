package main

import (
	"os"
)

type Model struct {
	view *View
	conf_name string
}

func die_if(err error, view *View) {
	if err != nil {
		view.log(ERROR, err.Error())
		os.Exit(1)
	}
}

func write_binary_file(buffer *[]byte, fname string, view *View) {
	f, err := os.Create(fname)
	die_if(err, view)
	defer f.Close()
	// hmm, look more closely at this 'defer'
	_, err = f.Write([]byte(*buffer))
	die_if(err, view)
}

func read_binary_file(buffer *[]byte, fname string, view *View) {
	f, err := os.Open(fname)
	die_if(err, view)
	defer f.Close()
	// hmm, look more closely at this 'defer'
	_, err = f.Read([]byte(*buffer))
	die_if(err, view)
}

func size_binary_file(fname string, view *View) uint64 {
	f, err := os.Stat(fname)
	die_if(err, view)
	return uint64(f.Size())
}

func zero_buffer(buffer *[]byte) {
	for key := range *buffer {
		(*buffer)[key] = 0
	}
}
