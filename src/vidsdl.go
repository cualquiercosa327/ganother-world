//SDL video implementation
//TODO: implement multiple pages
package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	WINDOW_WIDTH  int32 = 320 * 3
	WINDOW_HEIGHT int32 = 200 * 3

	WIDTH  int32 = 320
	HEIGHT int32 = 200
)

type SDLRenderer struct {
	surface     *sdl.Surface
	renderer    *sdl.Renderer
	window      *sdl.Window
	videoAssets VideoAssets
	loadPalette int
	colors      [16]Color
	exitAppReq  bool

	workerPage int
}

func buildSDLRenderer() *SDLRenderer {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	window, err := sdl.CreateWindow("ganother world", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		WINDOW_WIDTH, WINDOW_HEIGHT, sdl.WINDOW_ALLOW_HIGHDPI)
	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
	if err != nil {
		panic(err)
	}
	renderer.SetLogicalSize(WIDTH, HEIGHT)
	renderer.Clear()
	renderer.Present()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	return &SDLRenderer{
		surface:  surface,
		window:   window,
		renderer: renderer,
	}
}

func (render *SDLRenderer) updateGamePart(videoAssets VideoAssets) {
	render.videoAssets = videoAssets
	render.colors = videoAssets.getPalette(0)
}

func (render SDLRenderer) drawString(color, posX, posY, stringId int) {
	text := getText(stringId)
	fmt.Printf(">VID: DRAWSTRING color:%d, x:%d, y:%d, text:%s\n", color, posX, posY, text)
	//setWorkPagePtr(buffer);?

	render.softwareVideo_SetColor(color)
	charPosX := int32(posX)
	charPosY := int32(posY)
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			charPosY += int32(FONT_HEIGHT)
			charPosX = int32(posX)
		} else {
			render.softwareVideo_DrawChar(charPosX, charPosY, text[i])
			charPosX += 8
		}
	}
}

func (render *SDLRenderer) drawShape(color, offset, zoom, posX, posY int) {
	render.videoAssets.videoPC = offset
	i := render.videoAssets.fetchByte()

	fmt.Printf(">VID: DRAWSHAPE i:%d, color:%d, offset:%d, x:%d, y:%d, zoom:%d\n", i, color, offset, posX, posY, zoom)

	if i >= 0xC0 {
		if color&0x80 > 0 {
			color = int(i & 0x3F)
		}
		render.softwareVideo_FillPolygon(color, zoom, posX, posY)
	} else {
		i &= 0x3F
		if i == 1 {
			fmt.Printf("drawShape INVALID! (1 != 2)\n")
		} else if i == 2 {
			render.softwareVideo_DrawShapeParts(zoom, posX, posY)
		} else {
			fmt.Printf("drawShape INVALID! (%d != 2)\n", i)
		}
	}
}

func (render SDLRenderer) fillPage(page, color int) {
	fmt.Println(">VID: FILLPAGE", page, color)
	render.softwareVideo_SetColor(color)
	render.softwareVideo_FillBuffer()
}

func (render SDLRenderer) copyPage(src, dst, vscroll int) {
	fmt.Println(">VID: COPYPAGE", src, dst, vscroll)
}

// blit
func (render *SDLRenderer) updateDisplay(page int) {
	fmt.Println(">VID: UPDATEDISPLAY", page)

	if render.loadPalette != 0xFF {
		fmt.Println(">VID: UPDATEPAL", render.loadPalette)
		//render.colors = render.videoAssets.getPalette(render.loadPalette)
		render.loadPalette = 0xFF
	}

	//TODO why is this needed?
	render.renderer.Present()
}

func (render SDLRenderer) setWorkPagePtr(page int) {
	fmt.Println(">VID: SETWORKPAGEPTR", page)
	render.updateWorkerPage(page)
}

func (render *SDLRenderer) setPalette(index int) {
	//render.loadPalette = index >> 8
	//TODO move this to updateDisplay and remove me
	render.colors = render.videoAssets.getPalette(index >> 8)
	fmt.Println(">VID: SETPALETTE", index>>8)
}

func (render *SDLRenderer) mainLoop() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			render.exitAppReq = true
			fmt.Println(">render.exitAppReq", render.exitAppReq)
		case *sdl.KeyboardEvent:
			fmt.Println(">render.exitAppReq2", t.Keysym.Sym)
			if t.Keysym.Sym == sdl.K_ESCAPE && t.State == 1 {
				render.exitAppReq = true
			}
		}
	}
}

func (render *SDLRenderer) shutdown() {
	render.window.Destroy()
	sdl.Quit()
}

func (render SDLRenderer) exitRequested(frameCount int) bool {
	return render.exitAppReq
}

// ----

func (render *SDLRenderer) updateWorkerPage(page int) {
	if page >= 0 && page <= 3 {
		render.workerPage = page
		return
	}
	switch page {
	case 0xFF:
		render.workerPage = 2
	case 0xFE:
		render.workerPage = 1
	default:
		render.workerPage = 0
		fmt.Println("updateWorkerPage != [0,1,2,3,0xFF,0xFE] ==", page)
	}
}

func (render SDLRenderer) softwareVideo_SetColor(color int) {
	col := render.colors[color]
	fmt.Println(">VID: SETCOLOR", color, col)
	render.renderer.SetDrawColor(col.r, col.g, col.g, 255)
}

func (render SDLRenderer) softwareVideo_FillBuffer() {
	rect := sdl.Rect{0, 0, WIDTH, HEIGHT}
	render.renderer.FillRect(&rect)
}

func (render SDLRenderer) softwareVideo_DrawChar(posX, posY int32, char byte) {
	ofs := 8 * (int32(char) - 0x20)
	for j := int32(0); j < 8; j++ {
		ch := FONT[ofs+j]
		for i := int32(0); i < 8; i++ {
			if ch&(1<<(7-i)) > 0 {
				render.renderer.DrawPoint(posX+i, posY+j)
			}
		}
	}
}

func (render *SDLRenderer) softwareVideo_FillPolygon(color, zoom, posX, posY int) {
	fmt.Printf(">VID: FILLPOLYGON color:%d, x:%d, y:%d, zoom:%d\n", color, posX, posY, zoom)

	bbw := int(render.videoAssets.fetchByte()) * zoom / 64
	bbh := int(render.videoAssets.fetchByte()) * zoom / 64

	x1 := posX - bbw/2
	x2 := posX + bbw/2
	y1 := posY - bbh/2
	y2 := posY + bbh/2

	if x1 > 319 || x2 < 0 || y1 > 199 || y2 < 0 {
		fmt.Println(">VID: FILLPOLYGON INVALID")
		return
	}

	col := render.colors[color%16]
	numVertices := int(render.videoAssets.fetchByte())

	var vx, vy = make([]int16, numVertices), make([]int16, numVertices)
	for i := 0; i < numVertices; i++ {
		vx[i] = int16(x1 + int(render.videoAssets.fetchByte())*zoom/64)
		vy[i] = int16(y1 + int(render.videoAssets.fetchByte())*zoom/64)
	}

	fmt.Println(">VID: FILLPOLYGON", numVertices, vx, vy)
	gfx.FilledPolygonColor(render.renderer, vx, vy, sdl.Color{col.r, col.g, col.g, 255})
}

func (render SDLRenderer) softwareVideo_DrawShapeParts(zoom, posX, posY int) {
	x := posX - int(render.videoAssets.fetchByte())*zoom/64
	y := posY - int(render.videoAssets.fetchByte())*zoom/64
	n := int16(render.videoAssets.fetchByte())
	fmt.Printf(">VID: DRAWSHAPEPARTS x:%d, y:%d, n:%d\n", x, y, n)

	for ; n >= 0; n-- {
		off := render.videoAssets.fetchWord()
		_x := x + int(render.videoAssets.fetchByte())*zoom/64
		_y := y + int(render.videoAssets.fetchByte())*zoom/64

		fmt.Printf(">VID: DRAWSHAPEPARTS off:%d at %d/%d\n", off, _x, _y)

		var color uint16 = 0xFF
		if off&0x8000 > 0 {
			readOfs := render.videoAssets.videoPC&0x7F
			b1 := render.videoAssets.cinematic[readOfs]
			color = uint16(b1)
			//TODO display head.. WTF is this?
			render.videoAssets.fetchWord()
		}
		off &= 0x7FFF

		oldVideoPc := render.videoAssets.videoPC
		render.drawShape(int(color), int(off*2), zoom, _x, _y)
		render.videoAssets.videoPC = oldVideoPc
	}
}
