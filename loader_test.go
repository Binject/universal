// +build windows

package universal

import (
	"io/ioutil"
	"log"
	"syscall"
	"testing"

	"github.com/Binject/debug/pe"
	"github.com/awgh/rawreader"
)

func Test_Windows_1(t *testing.T) {

	image, err := ioutil.ReadFile("test\\main.dll")

	loader, err := NewLoader(false)
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

func Test_Windows_2(t *testing.T) {

	image, err := ioutil.ReadFile("test\\main.dll")

	loader, err := NewLoader(false)
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

func Test_Windows_3(t *testing.T) {

	image, err := ioutil.ReadFile("test\\main.dll")

	loader, err := NewLoader(false)
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

/*
// To build main.dll:
// Filename: main.cpp
#include "main.h"
#define DLL_EXPORT
extern "C" {
    int Runme(int var) {
        return var;
    }
}

// Filename: main.h
#ifndef MAIN_DLL_H
#define MAIN_DLL_H
#if defined DLL_EXPORT
#define BUILDING_MAIN_DLL __declspec(dllexport)
#else
#define BUILDING_MAIN_DLL __declspec(dllimport)
#endif
extern "C" {
    int Runme(int var);
}
#endif

// To compile:
g++ -c main.cpp
g++ -shared -o main.dll main.o -W
*/

/*
	// This would work with a PEB write:
	mainDLL, err := syscall.LoadDLL("main")
	if err != nil {
		panic(err)
	}
	runMe, err := mainDLL.FindProc("Runme")
	if err != nil {
		panic(err)
	}
	output, _, err := runMe.Call(7)
	if err != nil {
		panic(err)
	}
	println(output)
*/

/*
	symbols, err := pefile.ImportedSymbols()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Imported Symbols: %+v\n", symbols)

	libs, err := pefile.ImportedLibraries()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Imported Libraries: %+v\n", libs)

	iml, err := bananaphone.InMemLoads()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded Libraries: %+v\n", iml)
*/
