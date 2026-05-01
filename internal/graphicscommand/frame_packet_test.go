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
	"image"
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/internal/graphics"
	"github.com/hajimehoshi/ebiten/v2/internal/graphicsdriver"
	"github.com/hajimehoshi/ebiten/v2/internal/shaderir"
)

type testGraphics struct {
	ops             []string
	setVerticesArgs []testSetVerticesCall
	drawArgs        []testDrawTrianglesCall
}

type testSetVerticesCall struct {
	vertices []float32
	indices  []uint32
}

type testDrawTrianglesCall struct {
	dst         graphicsdriver.ImageID
	srcs        [graphics.ShaderSrcImageCount]graphicsdriver.ImageID
	shader      graphicsdriver.ShaderID
	indexOffset int
	indexCount  int
}

func (g *testGraphics) Initialize() error { return nil }

func (g *testGraphics) Begin() error {
	g.ops = append(g.ops, "begin")
	return nil
}

func (g *testGraphics) End(present bool) error {
	g.ops = append(g.ops, "end")
	return nil
}

func (g *testGraphics) SetTransparent(transparent bool) {}

func (g *testGraphics) SetVertices(vertices []float32, indices []uint32) error {
	g.ops = append(g.ops, "set-vertices")
	g.setVerticesArgs = append(g.setVerticesArgs, testSetVerticesCall{
		vertices: append([]float32(nil), vertices...),
		indices:  append([]uint32(nil), indices...),
	})
	return nil
}

func (g *testGraphics) NewImage(width, height int) (graphicsdriver.Image, error) {
	return testGraphicsImage{id: 1}, nil
}

func (g *testGraphics) NewScreenFramebufferImage(width, height int) (graphicsdriver.Image, error) {
	return testGraphicsImage{id: 1}, nil
}

func (g *testGraphics) SetVsyncEnabled(enabled bool) {}

func (g *testGraphics) NeedsClearingScreen() bool { return false }

func (g *testGraphics) MaxImageSize() int { return 4096 }

func (g *testGraphics) NewShader(program *shaderir.Program) (graphicsdriver.Shader, error) {
	return testGraphicsShader{id: 1}, nil
}

func (g *testGraphics) DrawTriangles(dst graphicsdriver.ImageID, srcs [graphics.ShaderSrcImageCount]graphicsdriver.ImageID, shader graphicsdriver.ShaderID, dstRegions []graphicsdriver.DstRegion, indexOffset int, blend graphicsdriver.Blend, uniforms []uint32, fillRule graphicsdriver.FillRule) error {
	g.ops = append(g.ops, "draw")
	indexCount := 0
	for _, region := range dstRegions {
		indexCount += region.IndexCount
	}
	g.drawArgs = append(g.drawArgs, testDrawTrianglesCall{
		dst:         dst,
		srcs:        srcs,
		shader:      shader,
		indexOffset: indexOffset,
		indexCount:  indexCount,
	})
	return nil
}

type testGraphicsImage struct {
	id graphicsdriver.ImageID
}

func (i testGraphicsImage) ID() graphicsdriver.ImageID { return i.id }
func (i testGraphicsImage) Dispose()                   {}
func (i testGraphicsImage) ReadPixels(args []graphicsdriver.PixelsArgs) error {
	return nil
}
func (i testGraphicsImage) WritePixels(args []graphicsdriver.PixelsArgs) error {
	return nil
}

type testGraphicsShader struct {
	id graphicsdriver.ShaderID
}

func (s testGraphicsShader) ID() graphicsdriver.ShaderID { return s.id }
func (s testGraphicsShader) Dispose()                    {}

type markerCommand struct {
	executed *[]string
	name     string
}

func (c *markerCommand) String() string { return c.name }

func (c *markerCommand) Exec(commandQueue *commandQueue, graphicsDriver graphicsdriver.Graphics, indexOffset int) error {
	*c.executed = append(*c.executed, c.name)
	if g, ok := graphicsDriver.(*testGraphics); ok {
		g.ops = append(g.ops, c.name)
	}
	return nil
}

func (c *markerCommand) NeedsSync() bool { return false }

type dummyFrameLogger struct{}

func (dummyFrameLogger) FrameLogf(format string, args ...any) {}
func (dummyFrameLogger) Flush()                               {}

func newTestImage(id graphicsdriver.ImageID) *Image {
	return &Image{
		image:  testGraphicsImage{id: id},
		width:  16,
		height: 16,
		id:     int(id),
	}
}

func newTestShader(id graphicsdriver.ShaderID) *Shader {
	return &Shader{
		shader: testGraphicsShader{id: id},
		id:     int(id),
	}
}

func newTestDrawTrianglesCommand(dstID, srcID graphicsdriver.ImageID, shaderID graphicsdriver.ShaderID, indexCount int) *drawTrianglesCommand {
	dst := newTestImage(dstID)
	src := newTestImage(srcID)
	return &drawTrianglesCommand{
		dst:        dst,
		srcs:       [graphics.ShaderSrcImageCount]*Image{src},
		vertices:   make([]float32, 4*graphics.VertexFloatCount),
		blend:      graphicsdriver.BlendSourceOver,
		dstRegions: []graphicsdriver.DstRegion{{Region: image.Rect(0, 0, 4, 4), IndexCount: indexCount}},
		shader:     newTestShader(shaderID),
		fillRule:   graphicsdriver.FillRuleFillAll,
	}
}

func TestDrawFramePacketExec(t *testing.T) {
	driver := &testGraphics{}
	dtc1 := newTestDrawTrianglesCommand(1, 11, 21, 3)
	dtc2 := newTestDrawTrianglesCommand(2, 12, 22, 6)
	vertices := make([]float32, len(dtc1.vertices)+len(dtc2.vertices))
	indices := make([]uint32, dtc1.numIndices()+dtc2.numIndices())

	packet, nc, nv, ne := newDrawFramePacket([]command{dtc1, dtc2}, vertices, indices)
	if got, want := nc, 2; got != want {
		t.Fatalf("nc: got %d, want %d", got, want)
	}
	if got, want := nv, len(vertices); got != want {
		t.Fatalf("nv: got %d, want %d", got, want)
	}
	if got, want := ne, len(indices); got != want {
		t.Fatalf("ne: got %d, want %d", got, want)
	}

	if err := packet.Exec(driver); err != nil {
		t.Fatal(err)
	}

	if got, want := len(driver.setVerticesArgs), 1; got != want {
		t.Fatalf("len(driver.setVerticesArgs): got %d, want %d", got, want)
	}
	if got, want := len(driver.drawArgs), 2; got != want {
		t.Fatalf("len(driver.drawArgs): got %d, want %d", got, want)
	}
	if got, want := driver.drawArgs[0].dst, graphicsdriver.ImageID(1); got != want {
		t.Fatalf("first draw dst: got %d, want %d", got, want)
	}
	if got, want := driver.drawArgs[0].srcs[0], graphicsdriver.ImageID(11); got != want {
		t.Fatalf("first draw src: got %d, want %d", got, want)
	}
	if got, want := driver.drawArgs[0].shader, graphicsdriver.ShaderID(21); got != want {
		t.Fatalf("first draw shader: got %d, want %d", got, want)
	}
	if got, want := driver.drawArgs[0].indexOffset, 0; got != want {
		t.Fatalf("first draw indexOffset: got %d, want %d", got, want)
	}
	if got, want := driver.drawArgs[1].indexOffset, dtc1.numIndices(); got != want {
		t.Fatalf("second draw indexOffset: got %d, want %d", got, want)
	}
}

func TestCommandQueueFlushSplitsDrawPacketsAtNonDrawCommands(t *testing.T) {
	driver := &testGraphics{}
	executed := []string{}
	dtc1 := newTestDrawTrianglesCommand(1, 11, 21, 3)
	dtc2 := newTestDrawTrianglesCommand(2, 12, 22, 6)
	q := &commandQueue{
		commands: []command{
			dtc1,
			&markerCommand{executed: &executed, name: "marker"},
			dtc2,
		},
		vertices: make([]float32, len(dtc1.vertices)+len(dtc2.vertices)),
		indices:  make([]uint32, dtc1.numIndices()+dtc2.numIndices()),
	}

	if err := q.flush(driver, false, dummyFrameLogger{}); err != nil {
		t.Fatal(err)
	}

	if got, want := len(driver.setVerticesArgs), 2; got != want {
		t.Fatalf("len(driver.setVerticesArgs): got %d, want %d", got, want)
	}
	if got, want := len(driver.drawArgs), 2; got != want {
		t.Fatalf("len(driver.drawArgs): got %d, want %d", got, want)
	}
	if got, want := driver.drawArgs[0].indexOffset, 0; got != want {
		t.Fatalf("first draw indexOffset: got %d, want %d", got, want)
	}
	if got, want := driver.drawArgs[1].indexOffset, 0; got != want {
		t.Fatalf("second draw indexOffset: got %d, want %d", got, want)
	}
	if got, want := executed, []string{"marker"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("executed: got %v, want %v", got, want)
	}
	if got, want := driver.ops, []string{"begin", "set-vertices", "draw", "marker", "set-vertices", "draw", "end"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("driver ops: got %v, want %v", got, want)
	}
}
