// © Knug Industries 2010 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package stringbuffer

// Allocate a byte buffer to build strings from a set of
// smaller strings, if added content exceeds maximal buffer 
// size, the size of the stringbuffer doubles.

type StringBuffer struct {
    current, max int
    buffer       []byte
}


func New() *StringBuffer {
    s := new(StringBuffer)
    s.Clear()
    return s
}

func NewSize(size int) *StringBuffer {
    s := new(StringBuffer)
    s.current = 0
    s.max = size
    s.buffer = make([]byte, size)
    return s
}

func (s *StringBuffer) Add(more string) {

    if (len(more) + s.current) > s.max {

        s.resize()
        s.Add(more)

    } else {

        for i := 0; i < len(more); i++ {
            s.buffer[i+s.current] = more[i]
        }

        s.current += len(more)
    }
}

func (s *StringBuffer) AddBytes(b []byte) {
    s.Add(string(b))
}

func (s *StringBuffer) Clear() {
    s.buffer = make([]byte, 100)
    s.current = 0
    s.max = 100
}

func (s *StringBuffer) ClearSize(z int) {
    s.buffer = make([]byte, z)
    s.current = 0
    s.max = z
}

func (s *StringBuffer) Capacity() int {
    return s.max
}

func (s *StringBuffer) Len() int {
    return s.current
}

func (s *StringBuffer) String() string {
    slice := s.buffer[:s.current]
    return string(slice)
}

func (s *StringBuffer) Bytes() []byte {
    return s.buffer[:s.current]
}

func (s *StringBuffer) resize() {

    s.buffer = append(s.buffer, make([]byte, s.max*2)...)
    s.max += s.max * 2

}
