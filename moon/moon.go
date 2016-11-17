package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
)

var root = flag.String("root", ".", "Document root path")
var port = flag.Int("port", 26980, "Listener port")
var moon = flag.String("moon", "moon.png", "Moon image")

func main() {
	flag.Parse()
	log.Println("Starting tiny webserver for 7DTD moon texture...")
	log.Println("Document root path: ", *root)
	http.HandleFunc("/", moonHandler)
	log.Println("Listening on port: ", *port)
	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		log.Fatal("ERROR Listen and Serve: ", err)
	}
}

func moonHandler(w http.ResponseWriter, r *http.Request) {
	/*** Color of days map ***/
	colors := [3]color.RGBA{color.RGBA{255, 255, 255, 255}, color.RGBA{245, 215, 10, 255}, color.RGBA{240, 15, 65, 255}}
	daymap := [7]int{2, 0, 0, 0, 0, 1, 1}

	/*** Get parameters ***/
	day, err := strconv.Atoi(r.FormValue("day"))
	if err != nil {
		log.Println("Unable to read key ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	/*** Load moon image ***/
	file, err := os.Open(*root + fmt.Sprintf("%c", os.PathSeparator) + *moon)
	if err != nil {
		log.Println("Unable to read image: ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer file.Close()

	mask, err := png.Decode(file)
	if err != nil {
		log.Println("Unable to decode image: ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	/*** Create colored image ***/
	dst := image.NewRGBA(image.Rect(0, 0, 80, 80))
	draw.Draw(dst, dst.Bounds(), mask, image.ZP, draw.Src)
	src := image.NewRGBA(image.Rect(0, 0, 80, 80))
	draw.Draw(src, src.Bounds(), &image.Uniform{colors[daymap[day%7]]}, image.ZP, draw.Src)
	draw.DrawMask(dst, dst.Bounds(), src, image.ZP, mask, image.ZP, draw.Over)

	/*** Write image ***/
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, dst); err != nil {
		log.Println("Unable to encode image: ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("Unable to write image: ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
