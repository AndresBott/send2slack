package mbox

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

const startString = "From "

type handler struct {
	path    string
	fhandle *os.File
	fSize   int64
	cursor  int64
}

func New(fname string) (*handler, error) {

	// todo: deal with permission denied ( i.e wring user)

	absFname, err := filepath.Abs(fname)
	if err != nil {
		return nil, err
	}

	h := handler{
		path: absFname,
	}

	err = h.loadFile()
	if err != nil {
		return nil, err
	}

	return &h, nil

}

// load the file handler and gather file stats
func (h *handler) loadFile() error {
	// open the file
	fHandle, err := os.Open(h.path)
	if err != nil {
		return err
	}
	h.fhandle = fHandle

	// get the file size
	stat, err := fHandle.Stat()
	if err != nil {
		return err
	}
	filesize := stat.Size()
	h.fSize = filesize

	return nil
}

// ReadLastLine uses the opened file handler and iterates over the runes descending from the last in the file
// until a line break is found
func (h *handler) readLastLine() []byte {

	line := []byte{}
	filesize := h.fSize
	// don't try to read empty file
	if filesize == 0 {
		return line
	}
	for {

		h.cursor -= 1
		h.fhandle.Seek(h.cursor, io.SeekEnd)

		char := make([]byte, 1)
		h.fhandle.Read(char)

		if h.cursor != -1 && (char[0] == 10 || char[0] == 13) { // stop if we find a line
			break
		}

		line = append(line, char[0])

		if h.cursor == -filesize { // stop if we are at the begining of the file
			break
		}
	}
	return ReverseRune(line)
}

// Reset will put the pointer value to 0, back to the end of the file in order to read the mbox again
func (h *handler) Reset() {
	h.cursor = 0
}

// Read the last email in the mbox, to minimize potential concurrency issues, whe start a new file handler with
// new stat on every read/modify operation
func (h *handler) ReadLastMail(delete bool) ([][]byte, error) {

	err := h.loadFile()
	if err != nil {
		return nil, err
	}

	// read lines
	lines := [][]byte{}

	startByte := []byte(startString)
	startByleLen := len(startByte)

	for {

		l := h.readLastLine()

		if h.cursor <= -h.fSize { // stop if we are at the begining of the file
			break
		}

		lines = append(lines, l)

		// the line starts with const startString
		if len(l) >= startByleLen {
			if bytes.Compare(l[0:startByleLen], startByte) == 0 {
				break
			}
		}
	}

	// reverse the lines
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	if delete {
		err = os.Truncate(h.path, h.fSize+h.cursor)
		h.Reset()
		if err != nil {
			return nil, err
		}
	}

	err = h.fhandle.Close()
	if err != nil {
		return nil, err
	}
	return lines, nil
}

// HasMails checks if the file handler cursor has reached the beginning of the file,
// returning true as long as there are still bytes to read
func (h *handler) HasMails() bool {
	if h.fSize == 0 {
		return false
	}
	if h.cursor <= -h.fSize {
		return false
	}
	return true
}

// ReverseRune returns a string with the runes of s in reverse order.
// Invalid UTF-8 sequences, if any, will be reversed byte by byte.
func ReverseRune(b []byte) []byte {
	res := make([]byte, len(b))
	prevPos, resPos := 0, len(b)
	for pos := range b {
		resPos -= pos - prevPos
		copy(res[resPos:], b[prevPos:pos])
		prevPos = pos
	}
	copy(res[0:], b[prevPos:])
	return res
}
