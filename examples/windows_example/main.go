// +build windows

package main

import (
	"io/ioutil"
	"log"

	"github.com/Binject/universal"
)

// PtrSize - are we on a 32bit or 64bit system?
const PtrSize = 32 << uintptr(^uintptr(0)>>63)

func main() {
	var image []byte
	var err error

	if PtrSize == 64 {
		image, err = ioutil.ReadFile("..\\..\\test\\64\\main.dll")
	} else {
		image, err = ioutil.ReadFile("..\\..\\test\\32\\main.dll")
	}

	loader, err := universal.NewLoader()
	if err != nil {
		log.Fatal(err)
	}

	library, err := loader.LoadLibrary("main", &image)
	if err != nil {
		log.Fatal(err)
	}

	val, err := library.Call("Runme", 7)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v\n", val)
}
