package demo

const (
    EventRead    = uint32(1 << 1)
    EventWrite   = uint32(1 << 2)
    EventTimeout = uint32(1 << 3)
    EventError   = uint32(1 << 4)
)

type EventHandler interface {
    HandleInput(fd int) error
    HandleOutput(fd int) error
    HandleException(fd int) error
    HandleTimeout(id uint64) error
    GetHandle() int
    GetID() int
}