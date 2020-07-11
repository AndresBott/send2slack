package mbox

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

const startString = "From "

type Handler struct {
	path    string
	fhandle *os.File
	fSize   int64
	cursor  int64
}

func NewHandler(fname string) (*Handler, error) {

	absFname, err := filepath.Abs(fname)
	if err != nil {
		return nil, err
	}

	fhandle, err := os.Open(absFname)
	if err != nil {
		return nil, err
	}

	fStat, err := fhandle.Stat()
	if err != nil {
		return nil, err
	}

	h := Handler{
		path:    absFname,
		fhandle: fhandle,
		fSize:   fStat.Size(),
	}
	return &h, nil

}

// ReadLastLine uses the opened file handler and iterates over the runes descending from the last in the file
// until a line break is found
func (h *Handler) ReadLastLine() []byte {

	line := []byte{}
	filesize := h.fSize
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
func (h *Handler) Reset() {
	h.cursor = 0
}

// ReadLastMail reads the last email in the mbox file and returns it
func (h *Handler) ReadLastMail() [][]byte {

	filesize := h.fSize
	lines := [][]byte{}

	startByte := []byte(startString)
	startByleLen := len(startByte)

	for {
		l := h.ReadLastLine()

		if h.cursor <= -filesize { // stop if we are at the begining of the file
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
	return lines
}

// ConsumeLastMail will read the last mail and remove it from the mbox file
func (h *Handler) ConsumeLastMail() ([][]byte, error) {
	m := h.ReadLastMail()

	newSize := h.fSize + h.cursor
	err := os.Truncate(h.path, h.fSize+h.cursor)
	if err != nil {
		return nil, err
	}
	h.fSize = newSize

	return m, nil
}

// HasMails checks if the file handler cursor has reached the beginning of the file,
// returning true as long as there are still bytes to read
func (h *Handler) HasMails() bool {
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

func (h *Handler) Close() {
	h.fhandle.Close()
}
