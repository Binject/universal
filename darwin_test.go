// +build darwin

package universal

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/Binject/debug/macho"
)

const PtrSize = 32 << uintptr(^uintptr(0)>>63) // are we on a 32bit or 64bit system?

func Test_DyldToBundle_arm64_1(t *testing.T) {

	var image []byte
	var err error

	image, err = ioutil.ReadFile("test/64/main-arm64.dyld")
	if err != nil {
		t.Fatal(err)
	}
	machoFile, err := macho.NewFile(bytes.NewReader(image))
	if err != nil {
		t.Fatal(err)
	}
	machoFile.FileHeader.Type = macho.TypeBundle
	image, err = machoFile.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile("test/64/main-arm64.bundle", image, 0700)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Darwin_arm64_1(t *testing.T) {

	var image []byte
	var err error

	image, err = ioutil.ReadFile("test/64/main-arm64.dyld")
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatal(err)
	}

	library, err := loader.LoadLibrary("main", &image)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%+v\n", library.Exports)

	/*
		runmeProc := library.BaseAddress + uintptr(exports[0].VirtualAddress)
		var a []uintptr
		a = append(a, uintptr(7))

		r1, r2, errno := syscall.Syscall(runmeProc, uintptr(len(a)), a[0], 0, 0)
		log.Println(r1, r2, errno)
	*/
}
