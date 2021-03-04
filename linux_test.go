// +build linux

package universal

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/awgh/cppgo/asmcall/cdecl"
)

func Test_Linux_1(t *testing.T) {
	image, err := ioutil.ReadFile("test/64/main.so")

	loader, err := NewLoader()
	if err != nil {
		t.Fatal(err)
	}

	library, err := loader.LoadLibrary("main", &image)
	if err != nil {
		t.Fatal(err)
	}

	runmeProc, ok := library.FindProc("Runme")
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

func Test_Linux_2(t *testing.T) {
	image, err := ioutil.ReadFile("test/64/main.so")

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
