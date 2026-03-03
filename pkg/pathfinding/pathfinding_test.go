package pathfinding_test

import (
	"testing"

	"pixel-server/pkg/pathfinding"
)

func TestStraightHorizontalPath(t *testing.T) {
	l := pathfinding.ParseHeightmap("00000")
	from := *l.At(0, 0)
	to := *l.At(4, 0)
	opts := pathfinding.Options{AllowDiagonal: false}
	path := pathfinding.FindPath(l, from, to, opts)
	if path == nil {
		t.Fatal("expected path, got nil")
	}
	if len(path) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(path))
	}
	for i, s := range path {
		if s.X != int16(i+1) || s.Y != 0 {
			t.Fatalf("step %d: expected (%d,0), got (%d,%d)", i, i+1, s.X, s.Y)
		}
	}
}

func TestDiagonalPath(t *testing.T) {
	hm := "000\n000\n000"
	l := pathfinding.ParseHeightmap(hm)
	from := *l.At(0, 0)
	to := *l.At(2, 2)
	opts := pathfinding.Options{AllowDiagonal: true}
	path := pathfinding.FindPath(l, from, to, opts)
	if path == nil {
		t.Fatal("expected path, got nil")
	}
	if len(path) != 2 {
		t.Fatalf("expected 2 diagonal steps, got %d", len(path))
	}
}

func TestPathAroundWall(t *testing.T) {
	hm := "000\n0x0\n000"
	l := pathfinding.ParseHeightmap(hm)
	from := *l.At(0, 1)
	to := *l.At(2, 1)
	opts := pathfinding.Options{AllowDiagonal: true}
	path := pathfinding.FindPath(l, from, to, opts)
	if path == nil {
		t.Fatal("expected path around wall, got nil")
	}
	for _, s := range path {
		if s.X == 1 && s.Y == 1 {
			t.Fatal("path went through blocked tile (1,1)")
		}
	}
}

func TestStaircasePath(t *testing.T) {
	stair := pathfinding.ParseHeightmap("01234")
	sFrom := *stair.At(0, 0)
	sTo := *stair.At(3, 0)
	opts := pathfinding.Options{AllowDiagonal: false}
	stairPath := pathfinding.FindPath(stair, sFrom, sTo, opts)
	if stairPath == nil {
		t.Fatal("expected staircase path")
	}
	if len(stairPath) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(stairPath))
	}
}

func TestFlyingEntityBypassesBlocked(t *testing.T) {
	hm := "0x0"
	l := pathfinding.ParseHeightmap(hm)
	from := *l.At(0, 0)
	to := *l.At(2, 0)
	normalPath := pathfinding.FindPath(l, from, to, pathfinding.Options{AllowDiagonal: false})
	if normalPath != nil {
		t.Fatal("expected nil path for blocked middle")
	}
	flyPath := pathfinding.FindPath(l, from, to, pathfinding.Options{AllowDiagonal: false, Flying: true})
	if flyPath == nil {
		t.Fatal("expected path for flying entity")
	}
}

func TestNoPathExists(t *testing.T) {
	hm := "0x0"
	l := pathfinding.ParseHeightmap(hm)
	from := *l.At(0, 0)
	to := *l.At(2, 0)
	path := pathfinding.FindPath(l, from, to, pathfinding.Options{AllowDiagonal: false})
	if path != nil {
		t.Fatal("expected nil for impossible path")
	}
}

func TestBlockedDestination(t *testing.T) {
	hm := "00x"
	l := pathfinding.ParseHeightmap(hm)
	from := *l.At(0, 0)
	to := *l.At(2, 0)
	path := pathfinding.FindPath(l, from, to, pathfinding.Options{AllowDiagonal: false})
	if path != nil {
		t.Fatal("expected nil for blocked destination")
	}
}

func TestParseHeightmap(t *testing.T) {
	hm := "012\nxx3\n456"
	l := pathfinding.ParseHeightmap(hm)
	if l.Width != 3 || l.Height != 3 {
		t.Fatalf("expected 3x3, got %dx%d", l.Width, l.Height)
	}
	if l.At(0, 0).Z != 0 || l.At(1, 0).Z != 1 || l.At(2, 0).Z != 2 {
		t.Fatal("row 0 Z mismatch")
	}
	if l.At(0, 1).State != pathfinding.TileBlocked || l.At(1, 1).State != pathfinding.TileBlocked {
		t.Fatal("row 1 blocked mismatch")
	}
	if l.At(2, 1).Z != 3 {
		t.Fatalf("expected Z=3, got %v", l.At(2, 1).Z)
	}
}

func BenchmarkFindPath64x64(b *testing.B) {
	hm := ""
	for y := 0; y < 64; y++ {
		row := ""
		for x := 0; x < 64; x++ {
			row += "0"
		}
		if y > 0 {
			hm += "\n"
		}
		hm += row
	}
	l := pathfinding.ParseHeightmap(hm)
	from := *l.At(0, 0)
	to := *l.At(63, 63)
	opts := pathfinding.Options{AllowDiagonal: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := pathfinding.FindPath(l, from, to, opts)
		if path == nil {
			b.Fatal("expected path")
		}
	}
}
