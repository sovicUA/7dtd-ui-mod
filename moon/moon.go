package main

import (
	"bytes"
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"net/http"
	"strconv"
)

var root = flag.String("root", ".", "Document root path")
var port = flag.Int("port", 26980, "Listener port")

/*** Color of days map ***/
var colors = [3]color.NRGBA{color.NRGBA{255, 255, 255, 0}, color.NRGBA{255, 220, 0, 200}, color.NRGBA{255, 0, 0, 200}}
var daymap = [7]int{2, 0, 0, 0, 0, 1, 1}

func main() {
	flag.Parse()
	log.Println("Starting tiny webserver for 7DTD moon texture...")
	log.Println("Document root path: ", *root)
	http.HandleFunc("/moon", moonHandler)
	http.HandleFunc("/color", colorHandler)
	log.Println("Listening on port: ", *port)
	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		log.Fatal("ERROR Listen and Serve: ", err)
	}
}

type circle struct {
	P image.Point
	R int
	C color.NRGBA
	S bool
}

func (c *circle) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *circle) Bounds() image.Rectangle {
	return image.Rect(c.P.X-c.R, c.P.Y-c.R, c.P.X+c.R, c.P.Y+c.R)
}

func (c *circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.P.X), float64(y-c.P.Y), float64(c.R)
	if xx*xx+yy*yy < rr*rr {
		xr := math.Abs(float64(c.P.X - x))
		yr := math.Abs(float64(c.P.Y - y))
		ar := uint8((1 - math.Sqrt(xr*xr+yr*yr)/float64(c.R)) * float64(c.C.A))
		if c.S {
			ar = c.C.A
		}
		return color.NRGBA{c.C.R, c.C.G, c.C.B, ar}
	}
	return color.Alpha{0}
}

func colorHandler(w http.ResponseWriter, r *http.Request) {
	d, err := strconv.Atoi(r.FormValue("d"))
	if err != nil {
		log.Println("Unable to read key ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	c := colors[daymap[d%7]]
	cs := strconv.Itoa(int(c.R)) + "," + strconv.Itoa(int(c.G)) + "," + strconv.Itoa(int(c.B)) + "," + strconv.Itoa(int(c.A))
	
	/*** Write Response ***/
	if _, err := w.Write([]byte(cs)); err != nil {
		log.Println("Unable to write color: ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

func moonHandler(w http.ResponseWriter, r *http.Request) {
	/*** Get parameters ***/
	d, err := strconv.Atoi(r.FormValue("d"))
	if err != nil {
		log.Println("Unable to read key ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	t, err := strconv.ParseBool(r.FormValue("t"))
	if err != nil {
		t = true
	}
	/*** Create colored image ***/
	cXY := image.Point{11, 11}
	cR := 11
	dst := image.NewRGBA(image.Rect(0, 0, cXY.X+cR, cXY.Y+cR))
	draw.Draw(dst, dst.Bounds(), &circle{cXY, cR, colors[daymap[d%7]], t}, image.ZP, draw.Src)

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
