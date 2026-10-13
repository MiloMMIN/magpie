// check_rsrc fails when its argument, a Windows exe, has no .rsrc section:
// the icon, version and manifest never reached the binary. `go build` links
// rsrc_windows_<arch>.syso from the package folder only when it is there
// already (`make winres`), and nothing else puts it there — a checkout,
// worktree or copy has none, and the build still succeeds silently. Reading
// the PE section table needs no toolchain on the build machine.
//
//	go run build/check_rsrc.go dist/magpie-windows-amd64.exe
package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "use: check_rsrc <windows exe>")
		os.Exit(2)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "check-rsrc:", err)
		os.Exit(1)
	}
	defer f.Close()
	var mz [2]byte
	if _, err := f.ReadAt(mz[:], 0); err != nil || mz != [2]byte{'M', 'Z'} {
		fmt.Fprintln(os.Stderr, "check-rsrc: not a PE file:", os.Args[1])
		os.Exit(1)
	}
	var b4 [4]byte
	if _, err := f.ReadAt(b4[:], 0x3c); err != nil {
		fmt.Fprintln(os.Stderr, "check-rsrc: truncated PE header:", os.Args[1])
		os.Exit(1)
	}
	pe := int64(binary.LittleEndian.Uint32(b4[:]))
	// section table: PE sig (4) + file header (20) + optional header
	var nsec, optsz [2]byte
	f.ReadAt(nsec[:], pe+6)
	f.ReadAt(optsz[:], pe+20)
	sec := pe + 24 + int64(binary.LittleEndian.Uint16(optsz[:]))
	for i := int64(0); i < int64(binary.LittleEndian.Uint16(nsec[:])); i++ {
		var name [8]byte
		if _, err := f.ReadAt(name[:], sec+i*40); err == nil && string(name[:5]) == ".rsrc" {
			return
		}
	}
	fmt.Fprintln(os.Stderr, "check-rsrc: no .rsrc section in", os.Args[1])
	fmt.Fprintln(os.Stderr, "  the icon, version and manifest were not linked: run `make winres`")
	fmt.Fprintln(os.Stderr, "  before `go build`, or build through `make release-windows`")
	os.Exit(1)
}
