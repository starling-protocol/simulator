package visualizer

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"time"

	"github.com/starling-protocol/simulator"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	SCREEN_WIDTH  = 1500
	SCREEN_HEIGHT = 1000
)

var (
	whiteImage = ebiten.NewImage(3, 3)

	// whiteSubImage is an internal sub image of whiteImage.
	// Use whiteSubImage at DrawTriangles instead of whiteImage in order to avoid bleeding edges.
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

type LineStyle struct {
	Red       int
	Green     int
	Blue      int
	TempColor color.Color
	MainColor color.Color
	Width     int
	FadeRate  float64
}

func DefaultLineStyle() *LineStyle {
	return &LineStyle{
		Red:       0,
		Green:     0,
		Blue:      0,
		TempColor: color.RGBA{0, 0, 0, 255},
		MainColor: color.RGBA{0, 0, 0, 255},
		Width:     1,
		FadeRate:  0,
	}
}

func NewLineStyle(mainColor color.Color, tempColor color.Color, width int, fadeRate float64) *LineStyle {
	r, g, b, _ := uint32(0), uint32(0), uint32(0), uint32(0)
	if mainColor == nil {
		mainColor = color.Black
	}
	if tempColor != nil {
		r, g, b, _ = tempColor.RGBA()
	}
	return &LineStyle{
		Red:       int(uint8(r)) * 256,
		Green:     int(uint8(g)) * 256,
		Blue:      int(uint8(b)) * 256,
		MainColor: mainColor,
		TempColor: tempColor,
		Width:     width,
		FadeRate:  fadeRate,
	}
}

type NodeStyle struct {
	Red       int
	Green     int
	Blue      int
	MainColor color.Color
	TempColor color.Color
	Size      int
	FadeRate  float64
}

func NewNodeStyle(mainColor color.Color, tempColor color.Color, fadeRate float64) *NodeStyle {
	r, g, b, _ := uint32(0), uint32(0), uint32(0), uint32(0)
	if mainColor == nil {
		mainColor = color.White
	}
	if tempColor != nil {
		r, g, b, _ = tempColor.RGBA()
	}
	return &NodeStyle{
		Red:       int(uint8(r)) * 256,
		Green:     int(uint8(g)) * 256,
		Blue:      int(uint8(b)) * 256,
		MainColor: mainColor,
		TempColor: tempColor,
		Size:      1,
		FadeRate:  fadeRate,
	}
}

type Visualizer struct {
	s                *simulator.Simulator
	offsetX          float64
	offsetY          float64
	mouseClickPointX float64
	mouseClickPointY float64
	zoomFactor       float64
	speedFactor      float64
	frameCount       int
	record           bool
	drawObstacles    bool
	obstacles        [][]simulator.Coordinate
	lastUpdate       *time.Time
	time             time.Duration
}

func NewVisualizer(sim *simulator.Simulator, record bool, speedFactor float64, drawObstacles bool) *Visualizer {
	return &Visualizer{
		s:                sim,
		offsetX:          0,
		offsetY:          0,
		mouseClickPointX: 0,
		mouseClickPointY: 0,
		zoomFactor:       8,
		speedFactor:      speedFactor,
		frameCount:       0,
		record:           record,
		drawObstacles:    drawObstacles,
		obstacles:        [][]simulator.Coordinate{},
		lastUpdate:       nil,
		time:             0,
	}
}

func (v *Visualizer) Update() error {
	if ebiten.IsWindowBeingClosed() {
		if !v.s.IsTerminating() {
			v.s.Terminate()
		}
		os.Exit(0)
	}

	v.handleInput()

	if v.s.IsRunning() && v.speedFactor > 0 {
		// Calculate time
		now := time.Now()
		if v.lastUpdate == nil {
			v.lastUpdate = &now
		}

		deltaTime := now.Sub(*v.lastUpdate)
		deltaTime = time.Duration(float64(deltaTime) * v.speedFactor)

		v.time = v.time + deltaTime
		v.lastUpdate = &now

		v.s.Update(v.time)
	} else if v.speedFactor <= 0 {
		now := time.Now()
		v.lastUpdate = &now
	}

	return nil
}

func (v *Visualizer) Draw(screen *ebiten.Image) {
	if ebiten.IsWindowBeingClosed() {
		if !v.s.IsTerminating() {
			v.s.Terminate()
		}
		os.Exit(0)
	}

	screen.Fill(color.RGBA{255, 255, 255, 255})

	mouseX, mouseY := ebiten.CursorPosition()

	if v.drawObstacles {
		v.drawAllObstacles(screen)
	}

	for _, peer := range v.s.Peers() {
		// Draw connections
		var peer_a_x, peer_a_y = v.coordsToPixels(peer.NodeA().Coords())
		var peer_b_x, peer_b_y = v.coordsToPixels(peer.NodeB().Coords())
		var midpoint_x, midpoint_y = midpoint(peer_a_x, peer_a_y, peer_b_x, peer_b_y)

		lineStyle := v.getLineStyle(&peer, peer.NodeA())
		if lineStyle == nil {
			lineStyle = DefaultLineStyle()
		}
		col := color.RGBA{uint8(lineStyle.Red / 256), uint8(lineStyle.Green / 256), uint8(lineStyle.Blue / 256), 255}
		vector.StrokeLine(screen, float32(peer_a_x), float32(peer_a_y), float32(peer_b_x), float32(peer_b_y), float32(lineStyle.Width)*float32(v.zoomFactor)*0.1, col, true)

		lineStyle = v.getLineStyle(&peer, peer.NodeB())
		if lineStyle == nil {
			lineStyle = DefaultLineStyle()
		}
		col = color.RGBA{uint8(lineStyle.Red / 256), uint8(lineStyle.Green / 256), uint8(lineStyle.Blue / 256), 255}
		vector.StrokeLine(screen, float32(midpoint_x), float32(midpoint_y), float32(peer_b_x), float32(peer_b_y), float32(lineStyle.Width)*float32(v.zoomFactor)*0.1, col, true)
	}

	var mouseHoverNode *simulator.InternalNode = nil
	for _, node := range v.s.Nodes() {
		var x, y = v.coordsToPixels(node.Coords())
		nodeStyle := v.getColor(node)
		col := color.RGBA{0, 0, 255, 255}
		if nodeStyle != nil {
			// Draw nodes
			col = color.RGBA{uint8(nodeStyle.Red / 256), uint8(nodeStyle.Green / 256), uint8(nodeStyle.Blue / 256), 255}
		}
		radius := 0.5 * float32(v.zoomFactor)
		vector.DrawFilledCircle(screen, float32(x), float32(y), radius, col, true)
		if withinDist(mouseX, mouseY, int(x), int(y), float64(radius)) {
			mouseHoverNode = node
		}
	}

	if mouseHoverNode != nil {
		pm := (*mouseHoverNode.Data())["printmap"]
		if pm != nil {
			printMap, ok := pm.(map[string]string)
			if ok {
				str := ""
				for key, val := range printMap {
					str += fmt.Sprintf("%s: %s\n", key, val)
				}
				ebitenutil.DebugPrint(screen, str)
			}
		}
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("Speed: %f\nZoom: %f", v.speedFactor, v.zoomFactor))

	if v.record {
		out, err := os.Create(fmt.Sprintf("./img/frame-%09d.png", v.frameCount))
		if err != nil {
			panic(err)
		}
		defer out.Close()

		var opts jpeg.Options
		opts.Quality = 100

		err = png.Encode(out, screen)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	}
	v.frameCount++
}

func withinDist(x1 int, y1 int, x2 int, y2 int, dist float64) bool {
	diffX := math.Abs(float64(x1) - float64(x2))
	diffY := math.Abs(float64(y1) - float64(y2))
	return diffX*diffX+diffY*diffY < dist*dist
}

func (v *Visualizer) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return SCREEN_WIDTH, SCREEN_HEIGHT
}

// Starts the GUI using the given simulator, with the given speed factor.
// If record is true, the simulation is recorded
// If scenarioFilepath is an empty string, no obstacles are rendered
func StartGUI(s *simulator.Simulator, record bool, speedFactor float64, scenarioFilepath string) {

	if record {
		err := os.RemoveAll("./img/")
		if err != nil {
			panic(err)
		}
		err = os.MkdirAll("./img/", os.FileMode(0775))
		if err != nil {
			panic(err)
		}

		// ffmpeg -framerate 30 -pattern_type glob -i '*.png' -c:v libx264 -pix_fmt yuv420p out.mp4
	}

	visualizer := NewVisualizer(s, record, speedFactor, scenarioFilepath != "")

	whiteImage.Fill(color.White)

	if scenarioFilepath != "" {
		scenarioFile, err := os.Open(scenarioFilepath)
		if err != nil {
			panic(err)
		}

		byteValue, _ := io.ReadAll(scenarioFile)

		var result map[string]interface{}
		json.Unmarshal([]byte(byteValue), &result)

		obstacleList := [][]simulator.Coordinate{}

		// Get list of obstacles
		scenario := result["scenario"].(map[string]interface{})
		topography := scenario["topography"].(map[string]interface{})
		obstacles := topography["obstacles"].([]interface{})
		for _, obs := range obstacles {
			obstaclePoints := []simulator.Coordinate{}

			obstacle := obs.(map[string]interface{})
			shape := obstacle["shape"].(map[string]interface{})
			shapeType := shape["type"].(string)
			switch shapeType {
			case "RECTANGLE":
				x := shape["x"].(float64)
				y := shape["y"].(float64)
				width := shape["width"].(float64)
				height := shape["height"].(float64)
				obstaclePoints = append(obstaclePoints, simulator.Coordinate{X: x, Y: y})
				obstaclePoints = append(obstaclePoints, simulator.Coordinate{X: x + width, Y: y})
				obstaclePoints = append(obstaclePoints, simulator.Coordinate{X: x + width, Y: y + height})
				obstaclePoints = append(obstaclePoints, simulator.Coordinate{X: x, Y: y + height})
			case "POLYGON":
				points := shape["points"].([]interface{}) //TODO: Fix so that it does not only work for polygons, but for rectangles as well
				for _, p := range points {
					point := p.(map[string]interface{})
					obstaclePoints = append(obstaclePoints, simulator.Coordinate{X: float64(point["x"].(float64)), Y: float64(point["y"].(float64))})

				}
			}

			obstacleList = append(obstacleList, obstaclePoints)
		}

		// Find center
		attributes := topography["attributes"].(map[string]interface{})
		bounds := attributes["bounds"].(map[string]interface{})
		x := bounds["x"].(float64)
		y := bounds["y"].(float64)
		width := bounds["width"].(float64)
		height := bounds["height"].(float64)
		visualizer.offsetX = x - (width / 2)
		visualizer.offsetY = y + (height / 2)
		visualizer.zoomFactor = (width - x) / 110

		scenarioFile.Close()

		visualizer.obstacles = obstacleList
	}

	s.Start()

	ebiten.SetWindowSize(SCREEN_WIDTH, SCREEN_HEIGHT)
	ebiten.SetWindowTitle("BLE Simulator")
	ebiten.SetWindowClosingHandled(true)
	if err := ebiten.RunGame(visualizer); err != nil {
		panic("Error starting GUI!")
	}
}

func (v *Visualizer) drawAllObstacles(screen *ebiten.Image) {
	var path vector.Path

	for _, obstacle := range v.obstacles {
		for i, coord := range obstacle {
			if i == 0 {
				x, y := v.coordsToPixels(coord)
				path.MoveTo(float32(x), float32(y))
			} else {
				x, y := v.coordsToPixels(coord)
				path.LineTo(float32(x), float32(y))
			}
		}
		path.Close()
	}

	var vs []ebiten.Vertex
	var is []uint16
	vs, is = path.AppendVerticesAndIndicesForFilling(nil, nil)

	for i := range vs {
		vs[i].DstX = (vs[i].DstX + float32(0))
		vs[i].DstY = (vs[i].DstY + float32(0))
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = 0xaa / float32(0xff)
		vs[i].ColorG = 0xaa / float32(0xff)
		vs[i].ColorB = 0xaa / float32(0xff)
		vs[i].ColorA = 1
	}

	op := &ebiten.DrawTrianglesOptions{}
	op.AntiAlias = true
	op.FillRule = ebiten.EvenOdd
	screen.DrawTriangles(vs, is, whiteSubImage, op)
}

func (v *Visualizer) handleInput() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0) {
		x, y := ebiten.CursorPosition()
		v.mouseClickPointX = v.offsetX - float64(x)/v.zoomFactor
		v.mouseClickPointY = v.offsetY - float64(y)/v.zoomFactor
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButton0) {
		x, y := ebiten.CursorPosition()
		v.offsetX = v.mouseClickPointX + float64(x)/v.zoomFactor
		v.offsetY = v.mouseClickPointY + float64(y)/v.zoomFactor
	}
	dx, dy := ebiten.Wheel()
	if v.zoomFactor+dy*0.1*v.zoomFactor > 0 {
		v.zoomFactor += dy * 0.1 * v.zoomFactor
	}
	if v.speedFactor+0.05*float64(dx) > -0.01 {
		v.speedFactor += 0.05 * float64(dx)
	}
	if ebiten.IsKeyPressed(ebiten.KeyDigit0) {
		v.speedFactor = 0
	} else if ebiten.IsKeyPressed(ebiten.KeyDigit1) {
		v.speedFactor = 1
	} else if ebiten.IsKeyPressed(ebiten.KeyDigit2) {
		v.speedFactor = 2
	} else if ebiten.IsKeyPressed(ebiten.KeyDigit3) {
		v.speedFactor = 4
	} else if ebiten.IsKeyPressed(ebiten.KeyDigit4) {
		v.speedFactor = 8
	} else if ebiten.IsKeyPressed(ebiten.KeyDigit5) {
		v.speedFactor = 16
	} else if ebiten.IsKeyPressed(ebiten.KeyDigit6) {
		v.speedFactor = 32
	} else if ebiten.IsKeyPressed(ebiten.KeyDigit7) {
		v.speedFactor = 0.5
	} else if ebiten.IsKeyPressed(ebiten.KeyDigit8) {
		v.speedFactor = 0.25
	} else if ebiten.IsKeyPressed(ebiten.KeyDigit9) {
		v.speedFactor = 0.125
	}
}

func (v *Visualizer) coordsToPixels(coord simulator.Coordinate) (float64, float64) {
	var x = ((coord.X * v.zoomFactor) + (SCREEN_WIDTH / 2)) + v.offsetX*v.zoomFactor
	var y = ((-coord.Y * v.zoomFactor) + (SCREEN_HEIGHT / 2)) + v.offsetY*v.zoomFactor
	return x, y
}

func midpoint(a_x float64, a_y float64, b_x float64, b_y float64) (float64, float64) {
	var mid_x float64 = (a_x + b_x) / 2.0
	var mid_y float64 = (a_y + b_y) / 2.0
	return mid_x, mid_y
}

func (v *Visualizer) fadeNodeColor(nodeStyle *NodeStyle) {
	r, g, b, _ := nodeStyle.TempColor.RGBA()
	mr, mg, mb, _ := nodeStyle.MainColor.RGBA()
	fadeRate := nodeStyle.FadeRate

	if uint8(nodeStyle.Red/256) != uint8(mr) {
		newR := int(float64(int((int(uint8(r))-int(uint8(mr)))*256)) * (fadeRate * v.speedFactor))
		updated := nodeStyle.Red - newR
		if updated < 0 {
			nodeStyle.Red = 0
		} else if updated > 65535 {
			nodeStyle.Red = 65535
		} else {
			nodeStyle.Red -= newR
		}
	}

	if uint8(nodeStyle.Green/256) != uint8(mg) {
		newG := int(float64(int((int(uint8(g))-int(uint8(mg)))*256)) * (fadeRate * v.speedFactor))
		updated := nodeStyle.Green - newG
		if updated < 0 {
			nodeStyle.Green = 0
		} else if updated > 65535 {
			nodeStyle.Green = 65535
		} else {
			nodeStyle.Green -= newG
		}
	}

	if uint8(nodeStyle.Blue/256) != uint8(mb) {
		newB := int(float64(int((int(uint8(b))-int(uint8(mb)))*256)) * (fadeRate * v.speedFactor))
		updated := nodeStyle.Blue - newB
		if updated < 0 {
			nodeStyle.Blue = 0
		} else if updated > 65535 {
			nodeStyle.Blue = 65535
		} else {
			nodeStyle.Blue -= newB
		}
	}
}

func (v *Visualizer) getColor(node *simulator.InternalNode) *NodeStyle {
	ns := (*node.Data())["nodestyle"]
	if ns != nil {
		nodeStyle, ok := ns.(*NodeStyle)
		if ok {
			if nodeStyle.TempColor == nil {
				nodeStyle.TempColor = nodeStyle.MainColor
			}
			if nodeStyle.FadeRate > 0 {
				v.fadeNodeColor(nodeStyle)
			}
			return nodeStyle
		}
	}
	return nil
}

func (v *Visualizer) fadeLineColor(lineStyle *LineStyle) {
	r, g, b, _ := lineStyle.TempColor.RGBA()
	mr, mg, mb, _ := lineStyle.MainColor.RGBA()
	fadeRate := lineStyle.FadeRate

	if uint8(lineStyle.Red/256) != uint8(mr) {
		newR := int(float64(int((int(uint8(r))-int(uint8(mr)))*256)) * (fadeRate * v.speedFactor))
		updated := lineStyle.Red - newR
		if updated < 0 {
			lineStyle.Red = 0
		} else if updated > 65535 {
			lineStyle.Red = 65535
		} else {
			lineStyle.Red -= newR
		}
	}

	if uint8(lineStyle.Green/256) != uint8(mg) {
		newG := int(float64(int((int(uint8(g))-int(uint8(mg)))*256)) * (fadeRate * v.speedFactor))
		updated := lineStyle.Green - newG
		if updated < 0 {
			lineStyle.Green = 0
		} else if updated > 65535 {
			lineStyle.Green = 65535
		} else {
			lineStyle.Green -= newG
		}
	}

	if uint8(lineStyle.Blue/256) != uint8(mb) {
		newB := int(float64(int((int(uint8(b))-int(uint8(mb)))*256)) * (fadeRate * v.speedFactor))
		updated := lineStyle.Blue - newB
		if updated < 0 {
			lineStyle.Blue = 0
		} else if updated > 65535 {
			lineStyle.Blue = 65535
		} else {
			lineStyle.Blue -= newB
		}
	}

}

func (v *Visualizer) getLineStyle(peer *simulator.InternalPeer, node *simulator.InternalNode) *LineStyle {
	ls := (*peer.Data(node))["linestyle"]
	if ls != nil {
		lineStyle, ok := ls.(*LineStyle)
		if ok {
			if lineStyle.TempColor == nil {
				lineStyle.TempColor = lineStyle.MainColor
			}
			if lineStyle.FadeRate > 0 {
				v.fadeLineColor(lineStyle)
			}
			return lineStyle
		}
	}
	return nil
}
