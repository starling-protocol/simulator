package starling_node

import (
	"encoding/json"
	"image/color"
	"strings"

	"github.com/starling-protocol/starling/device"
	"github.com/starling-protocol/starling/sync"

	"github.com/starling-protocol/simulator"
	"github.com/starling-protocol/simulator/visualizer"
)

func NetworkLayerColors(node *Node) {
	red := color.RGBA{250, 50, 50, 255}
	yellow := color.RGBA{250, 255, 0, 255}
	purple := color.RGBA{255, 0, 255, 255}
	orange := color.RGBA{255, 153, 51, 255}
	blue := color.RGBA{0, 0, 255, 255}

	//TODO: When multiple sessions are along a single edge, if one breaks, it will reset the color of the edge even though there is still another active session

	// RREQ
	node.AddScenario(
		EventScenario(LogEvent("network:packet:rreq:receive:")).
			OnEvent(ColorScenario(red, 0.01, red, 0.01, 0, 4)))
	node.AddScenario(
		EventScenario(LogEvent("network:packet:rreq:send:")).
			OnEvent(ColorScenario(red, 0.01, red, 0.01, 0, 4)))

	// TODO: Deal with broadcast

	// RREP
	node.AddScenario(
		EventScenario(LogEvent("network:packet:rrep:forward:")).
			OnEvent(ColorScenario(yellow, 0.0, yellow, 0.0, 4, 4)))
	node.AddScenario(
		EventScenario(LogEvent("network:packet:rrep:receive:")).
			OnEvent(ColorScenario(yellow, 0.0, yellow, 0.0, 4, 4)))

	// SESS
	node.AddScenario(
		EventScenario(LogEvent("network:send:sess:session:")).
			OnEvent(ColorScenario(purple, 0.05, purple, 0.05, 1, 5)))
	node.AddScenario(
		EventScenario(LogEvent("network:packet:sess:receive_packet:")).
			OnEvent(ColorScenario(purple, 0.05, purple, 0.05, 1, 4)))
	node.AddScenario(
		EventScenario(LogEvent("network:packet:sess:forward:")).
			OnEvent(ColorScenario(purple, 0.05, purple, 0.05, 1, 4)))

	// RERR
	node.AddScenario(
		EventScenario(LogEvent("network:packet:rerr:send:")).
			OnEvent(ColorScenario(blue, 0.0, color.Black, 0.0, 1, 4)).
			OnEvent(ColorScenario(orange, 0.01, orange, 0.01, 1, 4)))
	node.AddScenario(
		EventScenario(LogEvent("network:packet:rerr:receive:")).
			OnEvent(ColorScenario(blue, 0.0, color.Black, 0.0, 1, 4)).
			OnEvent(ColorScenario(orange, 0.01, orange, 0.01, 1, 4)))
}

type ScenarioColor struct {
	EmptyScenario
	nodeColor    color.Color
	nodeFadeRate float64
	lineColor    color.Color
	lineFadeRate float64
	lineWidth    int
	peerOffset   int
}

func ColorScenario(nodeColor color.Color, nodeFadeRate float64, lineColor color.Color, lineFadeRate float64, lineWidth int, peerOffset int) *ScenarioColor {
	return &ScenarioColor{
		nodeColor:    nodeColor,
		nodeFadeRate: nodeFadeRate,
		lineColor:    lineColor,
		lineFadeRate: lineFadeRate,
		lineWidth:    lineWidth,
		peerOffset:   peerOffset,
	}
}

func (s *ScenarioColor) OnLog(node *Node, message string) {
	if s.nodeFadeRate > 0.0 {
		node.colorNodeTemp(s.nodeColor, s.nodeFadeRate)
	} else {
		node.colorNode(s.nodeColor)
	}
	if s.peerOffset >= 0 {
		address := strings.Split(message, ":")[s.peerOffset]
		if s.lineFadeRate > 0.000001 {
			node.colorLineTempAddress(address, s.lineColor, s.lineFadeRate)
		} else {
			node.colorLineAddress(address, s.lineColor, s.lineWidth)
		}
	}
}

func (n *Node) colorNode(mainColor color.Color) {
	data := *n.sim.Data()
	ns := data["nodestyle"]
	if ns != nil {
		nodeStyle, ok := ns.(*visualizer.NodeStyle)
		if ok {
			nodeStyle.MainColor = mainColor // color.RGBA{250, 250, 0, 255}
			nodeStyle.TempColor = mainColor
			r, g, b, _ := mainColor.RGBA()
			nodeStyle.Red = int(uint8(r)) * 256
			nodeStyle.Green = int(uint8(g)) * 256
			nodeStyle.Blue = int(uint8(b)) * 256
			nodeStyle.FadeRate = 0
			return
		} else {
			panic("Error displaying colors")
		}
	}
	data["nodestyle"] = visualizer.NewNodeStyle(nil, mainColor, 0)
}

func (n *Node) colorNodeTemp(col color.Color, fadeRate float64) {
	data := *n.sim.Data()
	ns := data["nodestyle"]
	if ns != nil {
		nodeStyle, ok := ns.(*visualizer.NodeStyle)
		if ok {
			nodeStyle.TempColor = col
			r, g, b, _ := col.RGBA()
			nodeStyle.Red = int(uint8(r)) * 256
			nodeStyle.Green = int(uint8(g)) * 256
			nodeStyle.Blue = int(uint8(b)) * 256
			nodeStyle.FadeRate = fadeRate
			return
		} else {
			panic("Error displaying colors")
		}
	}

	data["nodestyle"] = visualizer.NewNodeStyle(color.RGBA{0, 0, 255, 255}, col, fadeRate)
}

func (n *Node) colorLineAddress(address string, col color.Color, width int) {
	peer, ok := n.peers[device.DeviceAddress(address)]
	if ok {
		n.colorLine(peer, col, width)
	}
}

func (n *Node) colorLineTempAddress(address string, col color.Color, fadeColor float64) {
	peer, ok := n.peers[device.DeviceAddress(address)]
	if ok {
		n.colorLineTemp(peer, col, fadeColor)
	}
}

func (n *Node) colorLine(peer simulator.Peer, col color.Color, width int) {
	d := peer.Data()
	if d == nil {
		return
	}
	data := *d
	ls := data["linestyle"]

	if ls != nil {
		lineStyle, ok := ls.(*visualizer.LineStyle)
		if ok {
			lineStyle.MainColor = col
			lineStyle.TempColor = col
			r, g, b, _ := col.RGBA()
			lineStyle.Red = int(uint8(r)) * 256
			lineStyle.Green = int(uint8(g)) * 256
			lineStyle.Blue = int(uint8(b)) * 256
			lineStyle.FadeRate = 0
			lineStyle.Width = width
			return
		} else {
			panic("Error displaying colors")
		}
	}

	data["linestyle"] = visualizer.NewLineStyle(col, col, width, 0)
}

func (n *Node) colorLineTemp(peer simulator.Peer, col color.Color, fadeRate float64) {
	d := peer.Data()
	if d == nil {
		return
	}
	data := *d
	ls := data["linestyle"]

	if ls != nil {
		lineStyle, ok := ls.(*visualizer.LineStyle)
		if ok {
			lineStyle.TempColor = col
			r, g, b, _ := col.RGBA()
			lineStyle.Red = int(uint8(r)) * 256
			lineStyle.Green = int(uint8(g)) * 256
			lineStyle.Blue = int(uint8(b)) * 256
			lineStyle.FadeRate = fadeRate
			return
		} else {
			panic("Error displaying colors")
		}
	}

	data["linestyle"] = visualizer.NewLineStyle(color.RGBA{0, 0, 0, 255}, col, 2, fadeRate)
}

var filterAllowAll = func(stateUpdate sync.Model) bool { return true }

func SyncModelDisplay(node *Node, contactID device.ContactID) {
	node.AddScenario(
		EventScenario(SyncStateChangedEvent(contactID, filterAllowAll)).
			OnEvent(SyncModelDisplayScenario()))
}

type ScenarioSyncModelDisplay struct {
	EmptyScenario
}

func SyncModelDisplayScenario() *ScenarioSyncModelDisplay {
	return &ScenarioSyncModelDisplay{}
}

func (s *ScenarioSyncModelDisplay) OnSyncStateChanged(node *Node, contact device.ContactID, stateUpdate sync.Model) {
	data := *node.sim.Data()
	pm := data["printmap"]

	if pm == nil {
		data["printmap"] = make(map[string]string)
		pm = data["printmap"]
	}

	printMap, ok := pm.(map[string]string)
	if ok {
		jsonState, err := json.MarshalIndent(stateUpdate, "", "\t")
		if err == nil {
			printMap["syncmodel"] = string(jsonState)
		}
	} else {
		panic("Error displaying sync model")
	}
}
