// Package rotatewriter contains additional tool for logging packages - RotateWriter Writer which implement normal fast smooth rotation
package rotatewriter

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// this file contains realizarion of Writer for logs which contains rotate capability

// RotateWriter is Writer with Rotate function to make correctly rotation of
type RotateWriter struct {
	Filename      string
	NumFiles      int
	dirpath       string
	file          *os.File
	writeMutex    sync.Mutex
	rotateMutex   sync.Mutex
	statusMap     sync.Map
	IsBuffered    bool
	FlushTimeout  time.Duration
	bufferedOut   *bufio.Writer
	BufferSize    int
	lastWriteTime time.Time
	ticker        *time.Ticker
}

// NewRotateWriter creates new instance make some checks there
// fileName: filename, must contain existing directory  file
// numfiles: 0 if no rotation at all - just reopen file on rotation. e.g. you would like use logrotate
// numfiles: >0 if rotation enabled
func NewRotateWriter(fileName string, numfiles int) (rw *RotateWriter, err error) {
	rw = &RotateWriter{Filename: fileName, NumFiles: numfiles, file: nil}
	err = rw.initDirPath()
	if nil != err {
		return nil, err
	}
	err = rw.openWriteFile()
	if nil != err {
		return nil, err
	}
	if 0 > numfiles {
		return nil, fmt.Errorf("numfiles must be 0 or more")
	}
	return rw, nil
}

// NewRotateBufferedWriter creates new buffered instance make some checks there
// fileName: filename, must contain existing directory  file
// numfiles: 0 if no rotation at all - just reopen file on rotation. e.g. you would like use logrotate
// numfiles: >0 if rotation enabled
// flush timeout - flush after timeout when there are no writes
// buffer size to work with
func NewRotateBufferedWriter(fileName string, numfiles int, flushTimeout time.Duration, bufferSize int) (rw *RotateWriter, err error) {
	rw = &RotateWriter{Filename: fileName, NumFiles: numfiles, file: nil, IsBuffered: true, FlushTimeout: flushTimeout, BufferSize: bufferSize}
	if flushTimeout == 0 {
		return nil, fmt.Errorf("flushTimeout must be not nil")
	}
	if bufferSize == 0 {
		return nil, fmt.Errorf("bufferSize must be non zero")
	}
	err = rw.initDirPath()
	if nil != err {
		return nil, err
	}
	err = rw.openWriteFile()
	if nil != err {
		return nil, err
	}
	if 0 > numfiles {
		return nil, fmt.Errorf("numfiles must be 0 or more")
	}
	rw.lastWriteTime = time.Now()
	rw.ticker = time.NewTicker(rw.FlushTimeout)
	// start flush ticker
	go func() {
		for t := range rw.ticker.C {
			func() {
				rw.writeMutex.Lock()
				defer rw.writeMutex.Unlock()
				duration := t.Sub(rw.lastWriteTime)
				if duration > rw.FlushTimeout && rw.bufferedOut != nil {
					// flush that shit
					rw.bufferedOut.Flush()
				}
			}()
		}
	}()
	return rw, nil
}

// initDirPath gets dir path from filename and init them
func (rw *RotateWriter) initDirPath() error {
	if rw.Filename == "" {
		return fmt.Errorf("Wrong log path")
	}
	rw.dirpath = filepath.Dir(rw.Filename)
	fileinfo, err := os.Stat(rw.dirpath)
	if err != nil {
		return err
	}
	if !fileinfo.IsDir() {
		return fmt.Errorf("Path to log file %s is not directory", rw.dirpath)
	}
	return nil
}

// openWriteFile warning - is not safe - use Lock unlock while work
func (rw *RotateWriter) openWriteFile() error {
	file, err := rw.openWriteFileInt()
	if err != nil {
		rw.file = nil
		return err
	}
	rw.file = file
	if rw.IsBuffered {
		rw.bufferedOut = bufio.NewWriterSize(rw.file, rw.BufferSize)
		rw.lastWriteTime = time.Now()
	}
	return nil
}

func (rw *RotateWriter) openWriteFileInt() (file *os.File, err error) {
	fileinfo, err := os.Stat(rw.Filename)
	newFile := false
	if err != nil {
		if os.IsNotExist(err) {
			newFile = true
			err = nil
		} else {
			return nil, err
		}
	}
	if fileinfo == nil {
		newFile = true
	} else {
		if fileinfo.IsDir() {
			return nil, fmt.Errorf("File %s is a directory", rw.Filename)
		}
	}
	if newFile {
		file, err = os.OpenFile(rw.Filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	} else {
		file, err = os.OpenFile(rw.Filename, os.O_APPEND|os.O_WRONLY, 0644)
	}
	return file, err
}

// CloseWriteFile use to close writer if you need
func (rw *RotateWriter) CloseWriteFile() error {
	rw.writeMutex.Lock()
	defer rw.writeMutex.Unlock()
	if rw.file == nil {
		return nil
	}
	if rw.IsBuffered {
		rw.bufferedOut.Flush()
		rw.bufferedOut = nil
	}
	err := rw.file.Close()
	rw.file = nil
	return err
}

// Write implements io.Writer
func (rw *RotateWriter) Write(p []byte) (n int, err error) {
	rw.writeMutex.Lock()
	defer rw.writeMutex.Unlock()
	if rw.file == nil {
		return 0, fmt.Errorf("Error: no file was opened for work with")
	}
	if rw.IsBuffered {
		n, err = rw.bufferedOut.Write(p)
		rw.lastWriteTime = time.Now()
	} else {
		n, err = rw.file.Write(p)
	}
	return n, err
}

// RotationInProgress detects rotation is running now
func (rw *RotateWriter) RotationInProgress() bool {
	_, ok := rw.statusMap.Load("rotation")
	return ok
}

// Rotate rotates file
func (rw *RotateWriter) Rotate(ready func()) error {
	if _, ok := rw.statusMap.Load("rotation"); ok {
		// rotation in progress - just prevent all fuckups
		return nil
	}
	rw.rotateMutex.Lock()
	defer rw.rotateMutex.Unlock()
	defer func() {
		if nil == ready {
			return
		}
		ready()
	}()
	rw.statusMap.Store("rotation", true)
	defer rw.statusMap.Delete("rotation")
	files, err := ioutil.ReadDir(rw.dirpath)
	if err != nil {
		return err
	}
	_, fileName := filepath.Split(rw.Filename)
	sl := make([]int, 0, rw.NumFiles)
	if rw.NumFiles > 0 {
	filesfor1:
		for _, fi := range files {
			if fi.IsDir() {
				return fmt.Errorf("Rotation problem: File %s is directory", fi.Name())
			}
			ext := filepath.Ext(fi.Name())
			if (fileName + ext) == fi.Name() {
				if ext == "" {
					continue filesfor1
				}
				ext = strings.Trim(ext, ".")
				num1, err := strconv.ParseInt(ext, 10, 64)
				num := int(num1)
				if (err != nil) && !os.IsNotExist(err) {
					continue filesfor1
				}
				if rw.NumFiles < num+1 { // unlink that shit
					err = os.Remove(path.Join(rw.dirpath, fi.Name()))
					if (err != nil) && !os.IsNotExist(err) {
						return err
					}
					continue filesfor1
				}
				sl = append(sl, num)
			}
		}
		sort.Slice(sl, func(i, j int) bool {
			return sl[i] > sl[j]
		})
		for _, num := range sl {
			err = os.Rename(
				path.Join(rw.dirpath, fileName+"."+strconv.FormatInt(int64(num), 10)),
				path.Join(rw.dirpath, fileName+"."+strconv.FormatInt(int64(num+1), 10)),
			)
			if err != nil {
				return err
			}
		}
		// may be we need errors here?
		os.Rename(rw.Filename, rw.Filename+".1")
	}
	renewFile := true
	if rw.NumFiles == 0 {
		// here may be one fail: file was not renamed
		_, err := os.Stat(rw.Filename)
		renewFile = os.IsNotExist(err)
	}
	// right way first open file - not to make program wait while Write()
	if renewFile { // if file was not deleted we really do not need reopen
		oldfile := rw.file
		oldBuff := rw.bufferedOut
		newfile, err := rw.openWriteFileInt()
		var newBuff *bufio.Writer
		newBuff = nil
		if rw.IsBuffered {
			newBuff = bufio.NewWriterSize(newfile, rw.BufferSize)
		}
		if err != nil {
			return err
		}
		func() { // just isolate rw.writeMutex work here
			rw.writeMutex.Lock()
			defer rw.writeMutex.Unlock()
			// now file is opened. Just make save switch of them
			rw.file = newfile
			if rw.IsBuffered {
				rw.bufferedOut = newBuff
			}
		}()
		if rw.IsBuffered {
			oldBuff.Flush()
		}
		oldfile.Close()
	}
	return nil
}
