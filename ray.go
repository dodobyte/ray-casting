package main

import (
	"bytes"
	"io/ioutil"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	W    = 1920
	H    = 1080
	Wall = 32
	FOV  = math.Pi / 3
)

var angle float64
var vel [2]float64
var pos = sdl.Point{30 * Wall, 20 * Wall}

var gameMap [][]byte
var maxLine = math.Hypot(W, H)
var distScreen = (W / 2.0) / math.Tan(FOV/2.0)

var quit bool
var renderer *sdl.Renderer

func distance(x, y int32) float64 {
	return math.Hypot(float64(pos.X-x), float64(pos.Y-y))
}

func intersectWall(rad float64) (bool, float64) {
	x1, y1 := pos.X, pos.Y
	x2 := x1 + int32(math.Cos(rad)*maxLine)
	y2 := y1 + int32(math.Sin(rad)*maxLine)
	min := maxLine
	for x := 0; x < len(gameMap); x++ {
		for y := 0; y < len(gameMap[0]); y++ {
			if gameMap[x][y] != '#' {
				continue
			}
			r := &sdl.Rect{int32(x * Wall), int32(y * Wall), Wall, Wall}
			X1, Y1, X2, Y2 := x1, y1, x2, y2
			if r.IntersectLine(&X1, &Y1, &X2, &Y2) {
				d := math.Min(distance(X1, Y1), distance(X2, Y2))
				if d < min {
					min = d
				}
			}
		}
	}
	return min < maxLine, min
}

func render() {
	renderer.SetDrawColor(135, 206, 235, 0)
	renderer.Clear()
	renderer.SetDrawColor(34, 139, 34, 0)
	floor := &sdl.Rect{0, H / 2, W, H / 2}
	renderer.FillRect(floor)
	for i := int32(0); i < W; i++ {
		rad := (angle - FOV/2) + FOV/W*float64(i)
		ok, d := intersectWall(rad)
		if ok {
			red := uint8(55)
			if d < 200 {
				red = 255 - uint8(d)
			}
			renderer.SetDrawColor(red, 0, 0, 0)
			size := int32(Wall / (d * math.Cos(rad-angle)) * distScreen)
			renderer.DrawLine(i, (H-size)/2, i, (H+size)/2)
		}
	}
	renderer.Present()
}

func input() {
	for ev := sdl.PollEvent(); ev != nil; ev = sdl.PollEvent() {
		switch t := ev.(type) {
		case *sdl.QuitEvent:
			quit = true
		case *sdl.KeyboardEvent:
			switch key := t.Keysym.Sym; key {
			case sdl.K_ESCAPE:
				quit = true
			case sdl.K_w:
				vel[0], vel[1] = 5, angle
			case sdl.K_s:
				vel[0], vel[1] = 5, angle+math.Pi
			case sdl.K_a:
				vel[0], vel[1] = 3, angle-math.Pi/2
			case sdl.K_d:
				vel[0], vel[1] = 3, angle+math.Pi/2
			}
			if t.Type == sdl.KEYUP {
				vel[0] = 0
			}
		}
	}
	x, _, _ := sdl.GetMouseState()
	switch {
	case x-W/2 < -5:
		angle -= 0.02
		sdl.WarpMouseGlobal(W/2, H/2)
	case x-W/2 > 5:
		angle += 0.02
		sdl.WarpMouseGlobal(W/2, H/2)
	}
	if vel[0] > 0 {
		pos.X += int32(vel[0] * math.Cos(vel[1]))
		pos.Y += int32(vel[0] * math.Sin(vel[1]))
	}
}

func loadMap(name string) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err)
	}
	rows := bytes.Split(data, []byte{0xA})
	w, h := len(rows[0]), len(rows)
	gameMap = make([][]byte, w)
	for x := 0; x < w; x++ {
		gameMap[x] = make([]byte, h)
		for y := 0; y < h; y++ {
			gameMap[x][y] = rows[y][x]
		}
	}
}

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	var undef int32 = sdl.WINDOWPOS_UNDEFINED
	var wndtype uint32 = sdl.WINDOW_FULLSCREEN
	wnd, err := sdl.CreateWindow("raycast", undef, undef, W, H, wndtype)
	if err != nil {
		panic(err)
	}
	defer wnd.Destroy()

	var flag uint32 = sdl.RENDERER_ACCELERATED | sdl.RENDERER_PRESENTVSYNC
	renderer, err = sdl.CreateRenderer(wnd, -1, flag)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	sdl.ShowCursor(sdl.DISABLE)
	loadMap("map.txt")

	for !quit {
		input()
		render()
	}
}
