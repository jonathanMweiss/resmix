package config

import (
	"fmt"
)

func CreateCascadeTopology(servers []string, numLayers int) *Topology {
	t := &Topology{
		Layers: make([]*Layer, 0, numLayers),
		Mixes:  make(map[string]*LogicalMix, numLayers*len(servers)),
	}

	for i := 0; i < numLayers; i++ {
		t.createLayer(servers, numLayers)
	}

	t.SetPredecessors()
	t.SetSuccessors()

	return t
}

func (x *Topology) createLayer(hostnames []string, totalLayers int) {
	layerNum := len(x.Layers)
	x.Layers = append(x.Layers, &Layer{})
	layer := x.Layers[layerNum]

	for i := range hostnames {
		mixPos := (i + layerNum) % len(hostnames)
		hostname := hostnames[mixPos]
		layer.LogicalMixes = append(layer.LogicalMixes, &LogicalMix{
			Hostname: hostname,
			Name:     fmt.Sprintf("m(%d,%d)", layerNum, mixPos),

			ServerIndex: int32(mixPos),
			Layer:       int32(layerNum),

			Predecessors: nil,
			Successors:   nil,
		})

		x.Mixes[layer.LogicalMixes[i].Name] = layer.LogicalMixes[i]
	}
}

const GenesisName = "GENESIS"

func (x *Topology) SetPredecessors() {
	for layerNum, layer := range x.Layers {
		for i, mix := range layer.LogicalMixes {
			if layerNum == 0 {
				mix.Predecessors = []string{GenesisName}

				continue
			}

			mix.Predecessors = append(mix.Predecessors, x.Layers[layerNum-1].LogicalMixes[i].Name)
		}
	}
}

func (x *Topology) SetSuccessors() {
	for layerNum, layer := range x.Layers {
		for i, mix := range layer.LogicalMixes {
			if layerNum == len(x.Layers)-1 {
				mix.Successors = []string{GenesisName}

				continue
			}

			mix.Successors = append(mix.Successors, x.Layers[layerNum+1].LogicalMixes[i].Name)
		}
	}
}
func (x *Topology) GetMix(name string) *LogicalMix {
	return x.Mixes[name]
}
