// +build windows

package universal

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/Binject/debug/pe"

	bananaphone "github.com/c-sto/BananaPhone/pkg/BananaPhone"
)

// Library - container struct for the DLL to load
type Library struct {
	Name   string
	Data   []byte
	file   *pe.File
	loaded bool
}

const (
	MEM_COMMIT                             = 0x001000
	MEM_RESERVE                            = 0x002000
	GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS = 0x00000004
	INFINITE                               = 0xFFFFFFFF
)

// LoadLibraries loads a set of libraries
// All dependencies must be provided or already loaded
func LoadLibraries(libraries []*Library) error {

	// retrieve already-loaded images from PEB
	inMemLoads, err := bananaphone.InMemLoads()
	if err != nil {
		return err
	}
	// parse all library images
	for _, lib := range libraries {
		pelib, err := pe.NewFile(bytes.NewReader(lib.Data))
		if err != nil {
			return err
		}
		lib.file = pelib
	}
	var locallyLoadedLibs []string

	lastUnloadedCount := -1
	// load loop
	for {
		unloadedCount := 0
		for _, lib := range libraries {
			if lib.loaded {
				continue // already loaded this one
			}
			requiredLibs, err := lib.file.ImportedLibraries() //todo: maybe the results of this should be cached somewhere
			if err != nil {
				return err
			}
			unsatisfiedCount := 0
			for _, needed := range requiredLibs {
				satisfied := false
				// check in-memory loads
				_, ok := inMemLoads[needed]
				if ok {
					satisfied = true
				} else {
					// check locally loaded libs
					for _, l := range locallyLoadedLibs {
						if l == needed {
							satisfied = true
							break
						}
					}
				}
				if !satisfied {
					unsatisfiedCount++
				}
			}
			if unsatisfiedCount == 0 {
				// all dependencies are satisfied, try to load this one
				err := LoadLibrary(lib.file, &lib.Data)
				if err == nil {
					locallyLoadedLibs = append(locallyLoadedLibs, lib.Name)
					lib.loaded = true
				} else {
					return err
				}
			} else {
				unloadedCount++
			}
		}
		if unloadedCount == 0 {
			break
		} else if unloadedCount == lastUnloadedCount {
			return errors.New("Dependencies cannot be satisfied with provided libraries")
		} else {
			lastUnloadedCount = unloadedCount
		}
	}
	return nil
}

// LoadLibrary - loads a single library to memory, without trying to check or load required imports
func LoadLibrary(pelib *pe.File, image *[]byte) error {
	pe64 := pelib.Machine == pe.IMAGE_FILE_MACHINE_AMD64
	var sizeOfImage uint32
	if pe64 {
		sizeOfImage = pelib.OptionalHeader.(*pe.OptionalHeader64).SizeOfImage
	} else {
		sizeOfImage = pelib.OptionalHeader.(*pe.OptionalHeader32).SizeOfImage
	}
	r, err := virtualAlloc(0, sizeOfImage, MEM_RESERVE, syscall.PAGE_READWRITE)
	if err != nil {
		return err
	}
	dst, err := virtualAlloc(r, sizeOfImage, MEM_COMMIT, syscall.PAGE_EXECUTE_READWRITE)
	if err != nil {
		return err
	}

	//perform base relocations
	pelib.Relocate(uint64(dst), image)

	//write to memory
	CopySections(pelib, dst)

	return nil
}

// CopySections - writes the sections of a PE image to the given base address in memory
func CopySections(pefile *pe.File, loc uintptr) error {
	for _, section := range pefile.Sections {
		fmt.Println("Writing:", fmt.Sprintf("%s %x %x", section.Name, loc, uint32(loc)+section.VirtualAddress))
		if section.Size == 0 {
			continue
		}
		d, err := section.Data()
		if err != nil {
			return err
		}
		dataLen := uint32(len(d))
		dst := uint64(loc) + uint64(section.VirtualAddress)
		buf := (*[^uint32(0)]byte)(unsafe.Pointer(uintptr(dst)))
		for index := uint32(0); index < dataLen; index++ {
			buf[index] = d[index]
		}
	}
	return nil
}

var (
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
	procVirtualAlloc = kernel32.MustFindProc("VirtualAlloc")
)

func virtualAlloc(addr uintptr, size, allocType, protect uint32) (uintptr, error) {
	r1, _, e1 := procVirtualAlloc.Call(
		addr,
		uintptr(size),
		uintptr(allocType),
		uintptr(protect))

	if int(r1) == 0 {
		return r1, os.NewSyscallError("VirtualAlloc", e1)
	}
	return r1, nil
}
