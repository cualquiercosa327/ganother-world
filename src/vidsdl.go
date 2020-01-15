//dummy video implementation, text output
package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	WINDOW_WIDTH  int32 = 640
	WINDOW_HEIGHT int32 = 480

	WIDTH  int32 = 320
	HEIGHT int32 = 200
)

type SDLRenderer struct {
	surface     *sdl.Surface
	renderer    *sdl.Renderer
	window      *sdl.Window
	videoAssets VideoAssets
	exitReq     bool

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

	/*	renderer, err := window.GetRenderer()
		if err != nil {
			panic(err)
		}*/
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
	if err != nil {
		panic(err)
	}
	renderer.SetLogicalSize(WIDTH, HEIGHT)
	renderer.Clear()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	//rect := sdl.Rect{0, 0, WIDTH, 100}
	renderer.SetDrawColor(250, 33, 110, 255)
	//renderer.FillRect(&rect)

	//renderer.Present()

	//	surface.FillRect(&rect, 0xff003300)
	//	window.UpdateSurface()

	return &SDLRenderer{
		surface:  surface,
		window:   window,
		renderer: renderer,
	}
}

func (render SDLRenderer) updateGamePart(videoAssets VideoAssets) {
	render.videoAssets = videoAssets
}

func (render SDLRenderer) drawString(color, posX, posY, stringId int) {
	text := getText(stringId)
	fmt.Printf(">VID: DRAWSTRING color:%d, x:%d, y:%d, text:%s\n", color, posX, posY, text)
	//TODO whats the color? index to palette?

	//setWorkPagePtr(buffer);?

	charPosX := int32(posX)
	charPosY := int32(posY)
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			charPosY += int32(FONT_HEIGHT)
			charPosX = int32(posX)
		} else {
			render.softwareVideo_DrawChar(color, charPosX, charPosY, text[i])
			charPosX += 8
		}
	}
	render.renderer.Present()
}

func (render SDLRenderer) drawShape(color, zoom, posX, posY int) {
	fmt.Printf(">VID: DRAWSHAPE color:%d, x:%d, y:%d, zoom:%d\n", color, posX, posY, zoom)
}

func (render SDLRenderer) fillPage(page, color int) {
	fmt.Println(">VID: FILLPAGE", page, color)
	//_graphics->clearBuffer(getPagePtr(page), color);
}

func (render SDLRenderer) copyPage(src, dst, vscroll int) {
	fmt.Println(">VID: COPYPAGE", src, dst, vscroll)
}

// blit
func (render SDLRenderer) updateDisplay(page int) {
	fmt.Println(">VID: UPDATEDISPLAY", page)
}

//TODO gimme a better name
func (render SDLRenderer) setDataBuffer(useSecondVideo bool, offset int) {
	fmt.Println(">VID: SETDATABUFFER", offset)
}

func (render SDLRenderer) setWorkPagePtr(page int) {
	fmt.Println(">VID: SETWORKPAGEPTR", page)
	render.updateWorkerPage(page)
}

func (render SDLRenderer) setPalette(index int) {
	fmt.Println(">VID: SETPALETTE", index>>8)
	//TODO	_vid->_nextPal = num >> 8
}

func (render *SDLRenderer) mainLoop() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			render.exitReq = true
			fmt.Println(">render.exitReq", render.exitReq)
		case *sdl.KeyboardEvent:
			if t.Type == 769 {
				render.exitReq = true
			}

		default:
			//fmt.Println("SDL EVENT", event)
		}
	}
}

func (render *SDLRenderer) shutdown() {
	render.window.Destroy()
	sdl.Quit()
}

func (render SDLRenderer) exitRequested(frameCount int) bool {
	return render.exitReq
}

//

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
		fmt.Println("Video::getPagePtr() p != [0,1,2,3,0xFF,0xFE] ==", page)
	}
}

func (render SDLRenderer) softwareVideo_DrawChar(color int, posX, posY int32, char byte) {
	ofs := 8 * (int32(char) - 0x20)
	for j := int32(0); j < 8; j++ {
		ch := FONT[ofs+j]
		for i := int32(0); i < 8; i++ {
			if ch&(1<<(7-i)) > 0 {
				render.renderer.DrawPoint(posX + i, posY + j)
			}
		}
	}
}