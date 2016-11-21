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

/*** circle (11(X), 11(Y), 11(R)) ***/
var cXY = image.Point{11, 11}
var cR = 11

/*** 
 * For blur circle: http://localhost:26980/moon?d={day}&t=false or http://localhost:26980/blur?d={day} 
 * For solid circle: http://localhost:26980/moon?d={day}&t=true or http://localhost:26980/moon?d={day} or http://localhost:26980/solid?d={day} 
 * ***/

func main() {
	flag.Parse()
	log.Println("Starting tiny webserver for 7DTD moon texture...")
	log.Println("Document root path: ", *root)
	http.HandleFunc("/blur", blurHandler)
	http.HandleFunc("/moon", moonHandler)
	http.HandleFunc("/solid", solidHandler)
	http.HandleFunc("/color", colorHandler)
	log.Println("Listening on port: ", *port)
	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		log.Fatal("ERROR Listen and Serve: ", err)
	}
}

/*** Circle interface ***/ 
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

/*** Get "day" parameters ***/
func getDay(r *http.Request) int {
	d, err := strconv.Atoi(r.FormValue("d"))
	if err != nil {
		log.Println("Unable to read key ", err)
		d = 0
	}
	return d
} 

/*** Get "type" parameters ***/
func getType(r *http.Request) bool {
	t, err := strconv.ParseBool(r.FormValue("t"))
	if err != nil {
		log.Println("Unable to read key ", err)
		t = true
	}
	return t
} 

/*** Write image ***/
func writeImg(w http.ResponseWriter, img image.Image) {
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
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

/*** Write color data only ***/ 
func colorHandler(w http.ResponseWriter, r *http.Request) {
	d := getDay(r)

	c := colors[daymap[d%7]]
	cs := strconv.Itoa(int(c.R)) + "," + strconv.Itoa(int(c.G)) + "," + strconv.Itoa(int(c.B)) + "," + strconv.Itoa(int(c.A))
	
	/*** Write Response ***/
	if _, err := w.Write([]byte(cs)); err != nil {
		log.Println("Unable to write color: ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

/*** Draw solid or blur circle | use t=false to draw blur circle ***/
func moonHandler(w http.ResponseWriter, r *http.Request) {
	d := getDay(r)
	t := getType(r)
	/*** Create colored image ***/
	dst := image.NewRGBA(image.Rect(0, 0, cXY.X+cR+1, cXY.Y+cR+1))
	draw.Draw(dst, dst.Bounds(), &circle{cXY, cR, colors[daymap[d%7]], t}, image.ZP, draw.Src)
	writeImg(w, dst)
}

/*** Draw blur circle ***/
func blurHandler(w http.ResponseWriter, r *http.Request) {
	d := getDay(r)
	/*** Create colored image ***/
	dst := image.NewRGBA(image.Rect(0, 0, cXY.X+cR+1, cXY.Y+cR+1))
	draw.Draw(dst, dst.Bounds(), &circle{cXY, cR, colors[daymap[d%7]], false}, image.ZP, draw.Src)
	writeImg(w, dst)
}

/*** Draw solid circle ***/
func solidHandler(w http.ResponseWriter, r *http.Request) {
	d := getDay(r)
	/*** Create colored image ***/
	dst := image.NewRGBA(image.Rect(0, 0, cXY.X+cR+1, cXY.Y+cR+1))
	draw.Draw(dst, dst.Bounds(), &circle{cXY, cR, colors[daymap[d%7]], true}, image.ZP, draw.Src)
	writeImg(w, dst)
}
