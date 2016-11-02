package main

import (
	"image"
	"log"
	"os"

	"github.com/mattn/drawcmd"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	drawcmd.Render(img)
}
