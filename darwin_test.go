// +build darwin

package universal

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/Binject/debug/macho"
	"github.com/awgh/cppgo/asmcall/cdecl"
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

	runmeProc, ok := library.FindProc("_Runme")
	if !ok {
		t.Fatal("FindProc did not find export Runme")
	}

	log.Printf("%x %x \n", runmeProc, library.BaseAddress)

	val, err := cdecl.Call(runmeProc, 7)
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("%+v\n", val)
}

func Test_Darwin_arm64_2(t *testing.T) {

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

	val, err := library.Call("Runme", 7)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("%+v\n", val)
}
