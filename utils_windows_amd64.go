package universal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"path/filepath"
	"strings"

	"github.com/Binject/debug/pe"
	"github.com/awgh/rawreader"
	"golang.org/x/sys/windows"
)

type addrExports struct {
	BaseAddr uint64
	Exports  *[]pe.Export
}

// UpdateIAT - reads the PEB and updates the Import Address Table (64bit version)
func UpdateIAT(pefile *pe.File) error {

	// have to parse all loaded DLL exports up-front to resolve Forwarder RVAs
	dllExports := make(map[string]addrExports)
	inMemLoads, err := InMemLoads()
	if err != nil {
		return err
	}
	for k, v := range inMemLoads {
		rr := rawreader.New(uintptr(v.BaseAddr), int(v.Size))
		p, err := pe.NewFileFromMemory(rr)
		if err != nil {
			return err
		}
		ex, err := p.Exports()
		if err != nil {
			return err
		}
		k = normalize(k)
		dllExports[k] = addrExports{BaseAddr: v.BaseAddr, Exports: &ex}
	}

	iatDir := pefile.IAT()
	log.Println(iatDir.VirtualAddress)

	importedSymbols, _ := pefile.ImportedSymbols()
	log.Println(importedSymbols, len(importedSymbols))

	ida, ds, sectionData, err := pefile.ImportDirectoryTable()
	if err != nil {
		return err
	}

	for _, dt := range ida {
		dt.DllName = normalize(dt.DllName)
		// skip non-distributable debug libraries
		// reference: https://docs.microsoft.com/en-us/cpp/c-runtime-library/crt-library-features?view=msvc-160
		if (len(dt.DllName) > 9 &&
			hashString(dt.DllName[0:9]) == 520068924 && /*VCRUNTIME*/
			dt.DllName[len(dt.DllName)-1:] == "D") ||
			hashString(dt.DllName) == 488540601 /*UCRTBASED*/ {
			continue
		}

		dllImage, ok := dllExports[dt.DllName]
		if !ok {
			return errors.New(dt.DllName + " is not loaded")
		}

		iatOffset := dt.FirstThunk - ds.VirtualAddress
		offset := dt.OriginalFirstThunk - ds.VirtualAddress
		for offset < uint32(len(*sectionData)) {
			va := binary.LittleEndian.Uint64((*sectionData)[offset : offset+8])
			if va == 0 {
				break
			}
			if va&0x8000000000000000 > 0 { // is Ordinal
				ordinal := uint32(va & 0x7fffffffffffffff)
				symbolAddr := findExportByOrdinal(dllImage.BaseAddr, dllImage.Exports, &dllExports, ordinal)
				if symbolAddr != 0 {
					b := new(bytes.Buffer)
					binary.Write(b, binary.LittleEndian, symbolAddr)
					copy((*sectionData)[iatOffset:iatOffset+8], b.Bytes())
				}
				log.Printf("%s->ORD(%d) %x %x\n", dt.DllName, ordinal, int64(dllImage.BaseAddr)+int64(symbolAddr), symbolAddr)
			} else {
				fn := getString(*sectionData, int(uint32(va)-ds.VirtualAddress+2))
				symbolAddr := findExportByName(dllImage.BaseAddr, dllImage.Exports, &dllExports, fn)
				if symbolAddr != 0 {
					b := new(bytes.Buffer)
					binary.Write(b, binary.LittleEndian, symbolAddr)
					copy((*sectionData)[iatOffset:iatOffset+8], b.Bytes())
				}
				log.Printf("%s->%s %x %x\n", dt.DllName, fn, int64(dllImage.BaseAddr)+int64(symbolAddr), symbolAddr)
			}
			offset += 8
			iatOffset += 8
		}
	}
	// sectionData is a copy of the actual section data, we have to copy the modifed version back in
	ds.Replace(bytes.NewReader(*sectionData), int64(len(*sectionData)))

	return nil
}

func resolveForwardedExport(allExports *map[string]addrExports, forward string) uint64 {
	idx := strings.Index(forward, ".")
	if idx > -1 && len(forward) > idx {
		lib := normalize(forward[:idx])
		symbol := forward[idx+1:]
		dllImage, ok := (*allExports)[lib]
		if ok {
			for _, ex := range *(dllImage.Exports) {
				if ex.Name == symbol {
					if ex.Forward != "" {
						panic("double forward")
						//return resolveForwardedExport(allExports, ex.Forward)
					}
					return dllImage.BaseAddr + uint64(ex.VirtualAddress)
				}
			}
		}
	}
	return 0
}

func findExportByName(base uint64, exports *[]pe.Export, allExports *map[string]addrExports, symbol string) uint64 {
	for _, s := range *exports {
		if s.Name == symbol {
			if s.Forward == "" {
				return base + uint64(s.VirtualAddress)
			}
			return resolveForwardedExport(allExports, s.Forward)
		}
	}
	return 0
}

func findExportByOrdinal(base uint64, exports *[]pe.Export, allExports *map[string]addrExports, ordinal uint32) uint64 {
	for _, s := range *exports {
		if s.Ordinal == ordinal {
			if s.Forward == "" {
				return base + uint64(s.VirtualAddress)
			}
			return resolveForwardedExport(allExports, s.Forward)
		}
	}
	return 0
}

func normalize(k string) string {
	k = strings.ToUpper(k)
	if len(k) > 4 && hashString(k[len(k)-4:]) == 174433490 /*.DLL*/ {
		k = k[:len(k)-4]
	}
	return k
}

// getString extracts a string from symbol string table.
func getString(section []byte, start int) string {
	if start < 0 || start >= len(section) {
		return ""
	}
	for end := start; end < len(section); end++ {
		if section[end] == 0 {
			return string(section[start:end])
		}
	}
	return ""
}

func hashString(str string) uint64 {
	p := uint64(131) // good for ASCII
	m := uint64(1e9 + 9)
	powerOfP := uint64(1)
	hashVal := uint64(0)
	for _, c := range str {
		hashVal = (hashVal + (uint64(c)+1)*powerOfP) % m
		powerOfP = (powerOfP * p) % m
	}
	return hashVal
}

// rest of this file stolen from https://github.com/C-Sto/BananaPhone

//stupidstring is the stupid internal windows definiton of a unicode string. I hate it.
type stupidstring struct {
	Length    uint16
	MaxLength uint16
	PWstr     *uint16
}

func (s stupidstring) String() string {
	return windows.UTF16PtrToString(s.PWstr)
}

//getModuleLoadedOrder returns the start address of module located at i in the load order. This might be useful if there is a function you need that isn't in ntdll, or if some rude individual has loaded themselves before ntdll.
func getModuleLoadedOrder(i int) (start uintptr, size uintptr, modulepath *stupidstring)

//GetModuleLoadedOrder returns the start address of module located at i in the load order. This might be useful if there is a function you need that isn't in ntdll, or if some rude individual has loaded themselves before ntdll.
func GetModuleLoadedOrder(i int) (start uintptr, size uintptr, modulepath string) {
	var badstring *stupidstring
	start, size, badstring = getModuleLoadedOrder(i)
	modulepath = badstring.String()
	return
}

// Image is an image
type Image struct {
	BaseAddr uint64
	Size     uint64
}

//InMemLoads returns a map of loaded dll basenames to current process offsets (aka images) in the current process. No syscalls are made.
func InMemLoads() (map[string]Image, error) {
	ret := make(map[string]Image)
	s, si, p := GetModuleLoadedOrder(0)
	start := p
	i := 1
	ret[strings.ToUpper(filepath.Base(p))] = Image{uint64(s), uint64(si)}
	for {
		s, si, p = GetModuleLoadedOrder(i)
		if p != "" {
			ret[strings.ToUpper(filepath.Base(p))] = Image{uint64(s), uint64(si)}
		}
		if p == start {
			break
		}
		i++
	}

	return ret, nil
}
