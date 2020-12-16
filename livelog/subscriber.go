package livelog

import "sync"

type subscriber struct {
	sync.Mutex

	handler chan *Line
	closec  chan struct{} // close channel
	closed  bool          // detect closed
}

func (s *subscriber) publish(line *Line) {
	select {
	case <-s.closec: // close channel
	case s.handler <- line:
	default:
	}
}

func (s *subscriber) close() {
	s.Lock()
	if !s.closed {
		// close channel closec
		close(s.closec)
		s.closed = true
	}

	s.Unlock()
}
