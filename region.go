package simulator

import "slices"

type RegionMap struct {
	regionMap        map[RegionCoord]*Region
	nodeRangeSquared float64
	regionSize       float64
}

func NewRegionMap(nodeRange float64) *RegionMap {
	return &RegionMap{
		regionMap:        make(map[RegionCoord]*Region),
		nodeRangeSquared: nodeRange * nodeRange,
		regionSize:       nodeRange,
	}
}

type RegionCoord struct {
	X int
	Y int
}

func (r *RegionMap) WithinRange(node *InternalNode) ([]*InternalNode, []*InternalNode) {
	coords := r.CoordToRegionCoord(node.coords)
	nodesWithinRange := []*InternalNode{}
	peersOutOfRange := []*InternalNode{}
	for _, x := range []int{-1, 0, 1} {
		for _, y := range []int{-1, 0, 1} {
			tmpCoords := RegionCoord{coords.X + x, coords.Y + y}
			_, found := r.regionMap[tmpCoords]
			if found {
				for _, nodeB := range r.regionMap[tmpCoords].Nodes {
					if node.internalID == nodeB.internalID {
						continue
					}
					if nodeDistSquared(node, nodeB) < r.nodeRangeSquared {
						nodesWithinRange = append(nodesWithinRange, nodeB)
					} else {
						if node.hasNeighbour(nodeB) {
							if nodeB.hasNeighbour(node) {
								peersOutOfRange = append(peersOutOfRange, nodeB)
							}
						}
					}
				}
			}
		}
	}

	return nodesWithinRange, peersOutOfRange
}

func (r *RegionMap) AddNode(node *InternalNode) {
	coord := r.CoordToRegionCoord(node.coords)
	_, found := r.regionMap[coord]
	if !found {
		r.regionMap[coord] = NewRegion(coord)
	}
	r.regionMap[coord].AddNode(node)
}

func (r *RegionMap) MoveNode(node *InternalNode, old Coordinate, new Coordinate) {
	oldCoords := r.CoordToRegionCoord(old)
	newCoord := r.CoordToRegionCoord(new)
	if oldCoords != newCoord {
		_, found := r.regionMap[oldCoords]
		if found {
			r.regionMap[oldCoords].RemoveNode(node)
		}
		_, found = r.regionMap[newCoord]
		if !found {
			r.regionMap[newCoord] = NewRegion(newCoord)
		}
		r.regionMap[newCoord].AddNode(node)
	}

}

func (r *RegionMap) CoordToRegionCoord(c Coordinate) RegionCoord {
	x := int(c.X / r.regionSize)
	y := int(c.Y / r.regionSize)
	return RegionCoord{x, y}
}

type Region struct {
	Coordinate RegionCoord
	Nodes      []*InternalNode
}

func NewRegion(c RegionCoord) *Region {
	return &Region{
		Coordinate: c,
		Nodes:      []*InternalNode{},
	}
}

func (r *Region) AddNode(node *InternalNode) {
	r.Nodes = append(r.Nodes, node)
}

func (r *Region) RemoveNode(node *InternalNode) {
	for i, n := range r.Nodes {
		if n == node {
			r.Nodes = slices.Delete(r.Nodes, i, i+1)
			return
		}
	}
}
