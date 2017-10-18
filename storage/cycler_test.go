package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCycler_add(t *testing.T) {
	cycler := Cycler{}
	cycler.add("foo")
	cycler.add("bar")

	assert.Equal(t, len(cycler.packages), 2)
}

func TestCycler_shiftByKeep(t *testing.T) {
	cycler := Cycler{
		packages: PackageList{
			Package{
				FileKey:   "p1",
				CreatedAt: time.Now(),
			},
			Package{
				FileKey:   "p2",
				CreatedAt: time.Now(),
			},
		},
	}
	cycler.add("p3")
	cycler.add("p4")
	cycler.add("p5")
	cycler.add("p6")

	pkg := cycler.shiftByKeep(2)
	assert.Equal(t, len(cycler.packages), 5)
	assert.Equal(t, pkg.FileKey, "p1")
	pkg = cycler.shiftByKeep(2)
	assert.Equal(t, len(cycler.packages), 4)
	assert.Equal(t, pkg.FileKey, "p2")
	pkg = cycler.shiftByKeep(4)
	assert.Equal(t, len(cycler.packages), 4)
	assert.Nil(t, pkg)
}
