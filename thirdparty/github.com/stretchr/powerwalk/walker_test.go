package powerwalk

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testFiles string = "./test_files"

func makeTestFiles(dirs, files int) {
	var counter int
	for i := 1; i < dirs+1; i++ {
		dir := fmt.Sprintf("%s/dir_%02d", testFiles, i)
		if err := os.MkdirAll(dir, 0777); err == nil {
			for j := 1; j < files+1; j++ {
				counter++
				filename := fmt.Sprintf("%s/file-%03d", dir, counter)
				ioutil.WriteFile(filename, []byte(fmt.Sprintf("This is file %d", counter)), 0777)
			}
		} else {
			panic(fmt.Sprintf("%s", err))
		}
	}
}
func deleteTestFiles() {
	os.RemoveAll("./test_files")
}

// BenchFilepathWalk uses the default Go implementation of filepath.Walk
func BenchmarkWalkFilepath(b *testing.B) {

	// max concurrency out
	runtime.GOMAXPROCS(runtime.NumCPU())

	b.StopTimer()
	makeTestFiles(10, 20)

	walkFunc := func(p string, info os.FileInfo, err error) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		filepath.Walk(testFiles, walkFunc)
	}

	b.StopTimer()
	deleteTestFiles()

}

// BenchmarkPowerwalk uses the power walker.
func BenchmarkPowerwalk(b *testing.B) {

	// max concurrency out
	runtime.GOMAXPROCS(runtime.NumCPU())

	b.StopTimer()
	makeTestFiles(10, 20)

	walkFunc := func(p string, info os.FileInfo, err error) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		Walk(testFiles, walkFunc)
	}

	b.StopTimer()
	deleteTestFiles()

}

func TestWalkFilepath(t *testing.T) {

	// max concurrency out
	runtime.GOMAXPROCS(runtime.NumCPU())

	makeTestFiles(10, 20)
	defer deleteTestFiles()

	seen := make(map[string]bool)
	walkFunc := func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			filename := path.Base(p)
			seen[filename] = true
		}
		return nil
	}

	assert.NoError(t, filepath.Walk(testFiles, walkFunc))

	// make sure everything was seen
	if assert.NotEqual(t, len(seen), 0, "Walker should visit at least one file.") {
		for k, v := range seen {
			assert.True(t, v, k)
		}
	}

}

func TestPowerWalk(t *testing.T) {

	// max concurrency out
	runtime.GOMAXPROCS(runtime.NumCPU())

	makeTestFiles(10, 20)
	defer deleteTestFiles()

	var seenLock sync.Mutex
	seen := make(map[string]bool)
	walkFunc := func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			filename := path.Base(p)
			seenLock.Lock()
			defer seenLock.Unlock()
			seen[filename] = true
		}
		return nil
	}

	assert.NoError(t, Walk(testFiles, walkFunc))

	// make sure everything was seen
	if assert.NotEqual(t, len(seen), 0, "Walker should visit at least one file.") {
		for k, v := range seen {
			assert.True(t, v, k)
		}
	}

}

/*
// This test is commented out as it takes an extremely long time.
func TestPowerWalkMassive(t *testing.T) {

	// max concurrency out
	runtime.GOMAXPROCS(runtime.NumCPU())

	rand.Seed(time.Now().UnixNano())

	makeTestFiles(200, 100)
	defer deleteTestFiles()

	count := 0
	total := 200 * 100

	var seenLock sync.Mutex
	seen := make(map[string]bool)
	walkFunc := func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			filename := path.Base(p)
			seenLock.Lock()
			seen[filename] = true
			count++
			seenLock.Unlock()

			// simulate some processing
			time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
			os.Stdout.Sync()
		}
		return nil
	}

	assert.NoError(t, Walk(testFiles, walkFunc))

	// make sure everything was seen
	if assert.NotEqual(t, len(seen), 0, "Walker should visit at least one file.") {
		for k, v := range seen {
			assert.True(t, v, k)
		}
	}

}
*/

func TestPowerWalkLimit(t *testing.T) {

	// max concurrency out
	runtime.GOMAXPROCS(runtime.NumCPU())

	makeTestFiles(10, 20)
	defer deleteTestFiles()

	var seenLock sync.Mutex
	seen := make(map[string]bool)
	walkFunc := func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			filename := path.Base(p)
			seenLock.Lock()
			defer seenLock.Unlock()
			seen[filename] = true
		}
		return nil
	}

	assert.NoError(t, WalkLimit(testFiles, walkFunc, 1))

	// make sure everything was seen
	if assert.NotEqual(t, len(seen), 0, "Walker should visit at least one file.") {
		for k, v := range seen {
			assert.True(t, v, k)
		}
	}

}

func TestPowerWalkLimitInvalidArgs(t *testing.T) {

	makeTestFiles(10, 20)
	defer deleteTestFiles()

	walkFunc := func(p string, info os.FileInfo, err error) error {
		return nil
	}
	assert.Panics(t, func() {
		WalkLimit(testFiles, walkFunc, 0)
	})

}

func TestPowerWalkLimitUselessThreadsDontBlock(t *testing.T) {

	makeTestFiles(10, 20)
	defer deleteTestFiles()

	walkFunc := func(p string, info os.FileInfo, err error) error {
		return nil
	}
	assert.NoError(t, WalkLimit(testFiles, walkFunc, 500))

}

func TestPowerWalkError(t *testing.T) {

	// max concurrency out
	runtime.GOMAXPROCS(runtime.NumCPU())

	makeTestFiles(10, 20)
	defer deleteTestFiles()

	theErr := errors.New("kaboom")
	var seenLock sync.Mutex
	seen := make(map[string]bool)
	walkFunc := func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			filename := path.Base(p)
			seenLock.Lock()
			defer seenLock.Unlock()
			if len(seen) > 20 {
				return theErr
			}
			seen[filename] = true
		}
		return nil
	}

	assert.Equal(t, Walk(testFiles, walkFunc), theErr)

	// make sure everything was seen
	if assert.NotEqual(t, len(seen), 0, "Walker should visit at least one file.") {
		for k, v := range seen {
			assert.True(t, v, k)
		}
	}

}
