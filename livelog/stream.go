package livelog

import (
	"context"
	"sync"
)

const bufferSize = 5000

type stream struct {
	sync.Mutex
	// return log one by one
	hist []*Line
	list map[*subscriber]struct{}
}

func newStream() *stream {
	return &stream{
		list: map[*subscriber]struct{}{},
	}
}

// write line to hist slice
func (s *stream) write(line *Line) error {
	s.Lock()
	s.hist = append(s.hist, line)
	for l := range s.list {
		l.publish(line)
	}
	// bufferSize log lines counter stored in memory
	if size := len(s.hist); size >= bufferSize {
		// remove the history line when capcacity is reached
		s.hist = s.hist[size-bufferSize:]
	}
	s.Unlock()
	return nil
}

// subscribe   in channl: Line struct, error
func (s *stream) subscribe(ctx context.Context) (<-chan *Line, <-chan error) {
	sub := &subscriber{
		// make a channel with <bufferSize> capacity size
		handler: make(chan *Line, bufferSize),
		// empty struct channel
		closec: make(chan struct{}),
	}
	// make error channel
	err := make(chan error)

	s.Lock()
	for _, line := range s.hist {
		sub.publish(line)
	}

	s.list[sub] = struct{}{}
	s.Unlock()

	go func() {
		defer close(err)
		select {
		case <-sub.closec:
		case <-ctx.Done():
			sub.close()
		}
	}()

	return sub.handler, err
}

// close clear the logs in stream
func (s *stream) close() error {
	s.Lock()
	defer s.Unlock()
	for sub := range s.list {
		delete(s.list, sub)
		sub.close()
	}

	return nil
}
