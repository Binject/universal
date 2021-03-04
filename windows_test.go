// +build windows

package universal

import (
	"fmt"
	"io/ioutil"
	"log"
	"syscall"
	"testing"

	"github.com/Binject/debug/pe"
	"github.com/awgh/rawreader"
)

const PtrSize = 32 << uintptr(^uintptr(0)>>63) // are we on a 32bit or 64bit system?

func Test_Windows_MSVC_1(t *testing.T) {

	var image []byte
	var err error

	if PtrSize == 64 {
		image, err = ioutil.ReadFile("test\\64\\TestDLL.dll")
	} else {
		image, err = ioutil.ReadFile("test\\32\\TestDLL.dll")
	}
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

	loadedDll, err := pe.NewFileFromMemory(rawreader.New(library.BaseAddress, int(^uint(0)>>1)))
	if err != nil {
		t.Fatal(err)
	}

	exports, err := loadedDll.Exports()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Exported Symbols from Loaded DLL: %+v\n", exports)

	runmeProc := library.BaseAddress + uintptr(exports[0].VirtualAddress)
	var a []uintptr
	a = append(a, uintptr(7))

	r1, r2, errno := syscall.Syscall(runmeProc, uintptr(len(a)), a[0], 0, 0)
	log.Println(r1, r2, errno)
}

func Test_Windows_gcc_2(t *testing.T) {

	var image []byte
	var err error

	if PtrSize == 64 {
		image, err = ioutil.ReadFile("test\\64\\main.dll")
	} else {
		image, err = ioutil.ReadFile("test\\32\\main.dll")
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatal(err)
	}

	_, err = loader.LoadLibrary("main", &image)
	if err != nil {
		t.Fatal(err)
	}

	runmeProc, ok := loader.FindProc("main", "Runme")
	if !ok {
		t.Fatal("FindProc did not find export Runme")
	}

	var a []uintptr
	a = append(a, uintptr(7))
	r1, r2, errno := syscall.Syscall(runmeProc, uintptr(len(a)), a[0], 0, 0)
	log.Println(r1, r2, errno)
}

func Test_Windows_gcc_3(t *testing.T) {

	var image []byte
	var err error

	if PtrSize == 64 {
		image, err = ioutil.ReadFile("test\\64\\main.dll")
	} else {
		image, err = ioutil.ReadFile("test\\32\\main.dll")
	}

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

	var a []uintptr
	a = append(a, uintptr(7))
	r1, r2, errno := syscall.Syscall(runmeProc, uintptr(len(a)), a[0], 0, 0)
	log.Println(r1, r2, errno)
}

func Test_Windows_gcc_4(t *testing.T) {

	var image []byte
	var err error

	if PtrSize == 64 {
		image, err = ioutil.ReadFile("test\\64\\main.dll")
	} else {
		image, err = ioutil.ReadFile("test\\32\\main.dll")
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

func Test_Windows_WrongArch_4(t *testing.T) {

	var image []byte
	var err error

	if PtrSize == 32 { // this SHOULD try to load the WRONG architecture DLL and error out on load!!!
		image, err = ioutil.ReadFile("test\\64\\TestDLL.dll")
	} else {
		image, err = ioutil.ReadFile("test\\32\\TestDLL.dll")
	}
	if err != nil {
		t.Fatal(err)
	}

	loader, err := NewLoader()
	if err != nil {
		t.Fatal(err)
	}

	_, err = loader.LoadLibrary("main", &image)
	if err == nil {
		t.Fatal("Did not error out when loading the wrong architecture!")
	}
	fmt.Println("Correctly returned error:", err)
}
