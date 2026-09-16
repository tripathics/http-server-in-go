package main

import (
	"fmt"
)

func flush_pending_bytes(src *[]byte, dest *[]byte) {
	copy(*dest, *src)
	*src = make([]byte, 0)
}

func main() {
	src := []byte("Neovim")
	dest := make([]byte, 8)
	fmt.Println(src)
	fmt.Println(dest)

	flush_pending_bytes(&src, &dest)
	fmt.Println("After copy")
	fmt.Println(src)
	fmt.Println(dest)
}
