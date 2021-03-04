// +build darwin

package main

import (
	"io/ioutil"
	"log"

	"github.com/Binject/universal"
)

func main() {
	var image []byte
	var err error

	image, err = ioutil.ReadFile("../../test/64/main-arm64.dyld")
	if err != nil {
		log.Fatal(err)
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
