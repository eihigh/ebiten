// Copyright 2026 The Ebitengine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package graphicscommand

import (
	"github.com/hajimehoshi/ebiten/v2/internal/graphics"
	"github.com/hajimehoshi/ebiten/v2/internal/graphicsdriver"
)

// drawFramePacket represents a packet of draw commands that can be executed
// independently from the game/update loop once the referenced GPU resources are ready.
type drawFramePacket struct {
	vertices []float32
	indices  []uint32
	draws    []drawTrianglesFramePacketCommand
}

type drawTrianglesFramePacketCommand struct {
	dst        graphicsdriver.ImageID
	srcs       [graphics.ShaderSrcImageCount]graphicsdriver.ImageID
	shader     graphicsdriver.ShaderID
	dstRegions []graphicsdriver.DstRegion
	blend      graphicsdriver.Blend
	uniforms   []uint32
	fillRule   graphicsdriver.FillRule
}

func newDrawFramePacket(commands []command, vertices []float32, indices []uint32) (drawFramePacket, int, int, int) {
	var nv, ne, nc int
	for _, c := range commands {
		dtc, ok := c.(*drawTrianglesCommand)
		if !ok {
			break
		}
		if nc > 0 && mustUseDifferentVertexBuffer(nv+dtc.numVertices()) {
			break
		}
		nv += dtc.numVertices()
		ne += dtc.numIndices()
		nc++
	}
	if nc == 0 {
		return drawFramePacket{}, 0, 0, 0
	}

	packet := drawFramePacket{
		vertices: vertices[:nv],
		indices:  indices[:ne],
		draws:    make([]drawTrianglesFramePacketCommand, nc),
	}
	for i, c := range commands[:nc] {
		packet.draws[i] = newDrawTrianglesFramePacketCommand(c.(*drawTrianglesCommand))
	}
	return packet, nc, nv, ne
}

func newDrawTrianglesFramePacketCommand(c *drawTrianglesCommand) drawTrianglesFramePacketCommand {
	var srcs [graphics.ShaderSrcImageCount]graphicsdriver.ImageID
	for i, src := range c.srcs {
		if src == nil {
			srcs[i] = graphicsdriver.InvalidImageID
			continue
		}
		srcs[i] = src.image.ID()
	}

	return drawTrianglesFramePacketCommand{
		dst:        c.dst.image.ID(),
		srcs:       srcs,
		shader:     c.shader.shader.ID(),
		dstRegions: c.dstRegions,
		blend:      c.blend,
		uniforms:   c.uniforms,
		fillRule:   c.fillRule,
	}
}

func (c drawTrianglesFramePacketCommand) numIndices() int {
	var n int
	for _, region := range c.dstRegions {
		n += region.IndexCount
	}
	return n
}

func (c drawTrianglesFramePacketCommand) Exec(graphicsDriver graphicsdriver.Graphics, indexOffset int) error {
	return graphicsDriver.DrawTriangles(c.dst, c.srcs, c.shader, c.dstRegions, indexOffset, c.blend, c.uniforms, c.fillRule)
}

func (p drawFramePacket) Exec(graphicsDriver graphicsdriver.Graphics) error {
	if len(p.draws) == 0 {
		return nil
	}
	if err := graphicsDriver.SetVertices(p.vertices, p.indices); err != nil {
		return err
	}
	indexOffset := 0
	for _, c := range p.draws {
		if err := c.Exec(graphicsDriver, indexOffset); err != nil {
			return err
		}
		indexOffset += c.numIndices()
	}
	return nil
}
