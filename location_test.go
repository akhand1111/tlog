package tlog

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocation(t *testing.T) {
	testLocationInside(t)
}

func testLocationInside(t *testing.T) {
	pc := Caller(0)
	name, file, line := pc.NameFileLine()
	assert.Equal(t, "tlog.testLocationInside", path.Base(name))
	assert.Equal(t, "location_test.go", filepath.Base(file))
	assert.Equal(t, 16, line)
}

func TestLocationShort(t *testing.T) {
	pc := Caller(0)
	assert.Equal(t, "location_test.go:24", pc.String())
}

func TestLocation2(t *testing.T) {
	func() {
		func() {
			l := Funcentry(0)

			assert.Equal(t, "location_test.go:30", l.String())
		}()
	}()
}

func TestLocationCropFileName(t *testing.T) {
	assert.Equal(t, "github.com/nikandfor/tlog/sub/module/file.go",
		cropFilename("/path/to/src/github.com/nikandfor/tlog/sub/module/file.go", "github.com/nikandfor/tlog/sub/module.(*type).method"))
	assert.Equal(t, "github.com/nikandfor/tlog/sub/module/file.go",
		cropFilename("/path/to/src/github.com/nikandfor/tlog/sub/module/file.go", "github.com/nikandfor/tlog/sub/module.method"))
	assert.Equal(t, "github.com/nikandfor/tlog/root.go", cropFilename("/path/to/src/github.com/nikandfor/tlog/root.go", "github.com/nikandfor/tlog.type.method"))
	assert.Equal(t, "github.com/nikandfor/tlog/root.go", cropFilename("/path/to/src/github.com/nikandfor/tlog/root.go", "github.com/nikandfor/tlog.method"))
	assert.Equal(t, "root.go", cropFilename("/path/to/src/root.go", "github.com/nikandfor/tlog.method"))
	assert.Equal(t, "sub/file.go", cropFilename("/path/to/src/sub/file.go", "github.com/nikandfor/tlog/sub.method"))
	assert.Equal(t, "root.go", cropFilename("/path/to/src/root.go", "tlog.method"))
	assert.Equal(t, "subpkg/file.go", cropFilename("/path/to/src/subpkg/file.go", "subpkg.method"))
	assert.Equal(t, "subpkg/file.go", cropFilename("/path/to/src/subpkg/file.go", "github.com/nikandfor/tlog/subpkg.(*type).method"))
}

func BenchmarkLocationCaller(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		Caller(0)
	}
}

func BenchmarkLocationNameFileLine(b *testing.B) {
	b.ReportAllocs()

	l := Caller(0)

	for i := 0; i < b.N; i++ {
		l.NameFileLine()
	}
}
