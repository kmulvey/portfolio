package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Kagami/go-avif"
	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
	log "github.com/sirupsen/logrus"
)

const originalImagesDir = "images/originals"
const fourKImagesDir = "images/4k"
const tenEightyImagesDir = "images/1080"
const smallEightyImagesDir = "images/small"

var resolutionToDir = map[uint]string{
	3840: fourKImagesDir,
	1920: tenEightyImagesDir,
	720:  smallEightyImagesDir,
}

type binFunc func(int, int) int
type encoder func(io.Writer, image.Image) error

func avifEncoder(w io.Writer, m image.Image) error {
	return avif.Encode(w, m, nil) // &avif.Options{Quality: avif.MaxQuality, Threads: 0})
}
func webpEncoder(w io.Writer, m image.Image) error {
	return webp.Encode(w, m, &webp.Options{Quality: 85})
}
func jpegEncoder(w io.Writer, m image.Image) error {
	return jpeg.Encode(w, m, &jpeg.Options{Quality: 85})
}

var extensionToencoder = map[string]encoder{
	".webp": webpEncoder,
	".avif": avifEncoder,
	".jpg":  jpegEncoder,
}

func main() {
	makeDirs(resolutionToDir)
	for _, originalImage := range listFiles(originalImagesDir) {
		convert(originalImage, resolutionToDir, extensionToencoder)
	}
}

func convert(from string, sizes map[uint]string, encoders map[string]encoder) {
	if filepath.Ext(from) != ".jpg" && filepath.Ext(from) != ".jpeg" {
		log.Fatal("input image ", from, " is not a jpg")
	}
	var filename = strings.TrimSuffix(filepath.Base(from), path.Ext(from))
	log.Info("converting ", from)

	var jpgFile, err = os.Open(from)
	HandleErr("jpg open", err)

	jpgData, err := jpeg.Decode(jpgFile)
	HandleErr("jpg decode", err)

	for extension, enc := range encoders {
		for width, dir := range sizes {
			var out, err = os.Create(filepath.Join(dir, filename+extension))
			HandleErr(extension+" create", err)

			var resized = resize.Resize(width, 0, jpgData, resize.Lanczos3)

			HandleErr(extension+" encode", enc(out, resized))
			HandleErr(extension+" close", out.Close())
			log.Info("created ", filepath.Join(dir, filename+extension))
		}
	}

	HandleErr("jpg close", jpgFile.Close())
}

func listFiles(root string) []string {
	var allFiles []string
	files, err := ioutil.ReadDir(root)
	HandleErr("readdir", err)

	for _, file := range files {
		if file.IsDir() {
			var subFiles = listFiles(path.Join(root, file.Name()))

			for _, subFile := range subFiles {
				allFiles = append(allFiles, subFile)
			}
		} else {
			allFiles = append(allFiles, path.Join(root, file.Name()))
		}
	}
	return allFiles
}

func makeDirs(dirs map[uint]string) {
	for _, dir := range dirs {
		var _, err = os.Stat(dir)
		if os.IsNotExist(err) {
			errDir := os.MkdirAll(fourKImagesDir, 0755)
			if errDir != nil {
				log.Fatal(err)
			}
		}
	}
}

func HandleErr(prefix string, err error) {
	if err != nil {
		log.Fatal(fmt.Errorf("%s: %w", prefix, err))
	}
}
