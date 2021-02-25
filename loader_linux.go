package universal

import (
	"bytes"
	"errors"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/Binject/debug/elf"
)

// LoadLibraryImpl - loads a single library to memory, without trying to check or load required imports
func LoadLibraryImpl(name string, image *[]byte) (*Library, error) {
	elflib, err := elf.NewFile(bytes.NewReader(*image))
	if err != nil {
		return nil, err
	}
	if elflib.FileHeader.Type != elf.ET_DYN {
		return nil, errors.New("Cannot load non-dynamic libraries")
	}
	exports, err := elflib.Exports()
	if err != nil {
		return nil, err
	}
	pageSize := uint64(unix.Getpagesize())
	var memSize uint64

	loads := make([]*elf.Prog, 0)
	for _, p := range elflib.Progs {
		if p.Type != elf.PT_LOAD || p.Filesz == 0 || p.Memsz == 0 {
			continue
		}
		highest := p.Memsz + p.Vaddr
		if highest > memSize {
			memSize = highest
		}
		loads = append(loads, p)
	}
	if memSize%pageSize != 0 { // round up to page size
		memSize = ((memSize / pageSize) * pageSize) + pageSize
	}
	baseBuf, err := unix.Mmap(-1, 0, int(memSize),
		unix.PROT_READ|unix.PROT_WRITE|unix.PROT_EXEC, unix.MAP_PRIVATE|unix.MAP_ANON)
	if err != nil {
		return nil, err
	}
	for _, p := range loads {
		copy(baseBuf[p.Vaddr:], (*image)[p.Off:p.Off+p.Filesz])
	}
	ex := make(map[string]uint64)
	for _, x := range exports {
		ex[x.Name] = x.VirtualAddress
	}
	library := Library{
		BaseAddress: uintptr(unsafe.Pointer(&baseBuf[0])),
		Exports:     ex,
	}
	return &library, nil
}
