// +build windows

package universal

import (
	"bytes"
	"io/ioutil"
	"log"
	"syscall"
	"testing"

	"github.com/Binject/debug/pe"
)

func Test_Windows_1(t *testing.T) {

	image, err := ioutil.ReadFile("test\\main.dll")

	pefile, err := pe.NewFile(bytes.NewReader(image))
	if err != nil {
		log.Fatal(err)
	}

	err = LoadLibrary(pefile, &image)
	if err != nil {
		t.Fatal(err)
	}

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

	/*
	   package main
	   import "syscall"
	   func main() {
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
	   }
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
}

/*
main.cpp
#include "main.h"
#define DLL_EXPORT
extern "C" {
    int Runme(int var) {
        return var;
    }
}





5:10
main.h
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
5:10
compile strings
g++ -c main.cpp
g++ -shared -o main.dll main.o -W

*/
