package livelog

import (
	"context"
	"errors"
	"sync"
)

var errStreamNotFound = errors.New("stream: not Found")

// Line each log item
type Line struct {
	Number    int    `json:"pos"`  //  Line number
	Message   string `json:"out"`  // line message
	Timestamp int64  `json:"time"` // line generated timestamp
}

// LogStreamInfo provides internal stream information. This can
// be used to monitor the number of registered streams and
// subscribers.
type LogStreamInfo struct {
	// Streams is a key-value pair where the key is the step
	// identifier, and the value is the count of subscribers
	// streaming the logs.
	Streams map[int64]int `json:"streams"`
}

// LogStream manages a live stream of logs. int64 for transaction ID
type LogStream interface {
	Create(context.Context, int64) error
	Delete(context.Context, int64) error
	Write(context.Context, int64, *Line) error
	Tail(context.Context, int64) (<-chan *Line, <-chan error)
	Info(context.Context) *LogStreamInfo
}

type streamer struct {
	sync.Mutex
	streams map[int64]*stream
}

// New return a new streamer
func New() LogStream {
	return &streamer{
		streams: make(map[int64]*stream),
	}
}

// Create method create a new stream
func (s *streamer) Create(ctx context.Context, id int64) error {
	s.Lock()
	s.streams[id] = newStream()
	s.Unlock()
	return nil
}

// Delete method delete stream from streams
func (s *streamer) Delete(ctx context.Context, id int64) error {
	s.Lock()
	stream, ok := s.streams[id]
	if ok {
		// delete stream id from the streams
		delete(s.streams, id)
	}
	s.Unlock()

	if !ok {
		return errStreamNotFound
	}

	return stream.close()
}

// Write method writer line to stream
func (s *streamer) Write(ctx context.Context, id int64, line *Line) error {
	s.Lock()
	stream, ok := s.streams[id]
	s.Unlock()

	if !ok {
		return errStreamNotFound
	}

	return stream.write(line)
}

func (s *streamer) Tail(ctx context.Context, id int64) (<-chan *Line, <-chan error) {
	s.Lock()
	stream, ok := s.streams[id]
	s.Unlock()

	if !ok {
		return nil, nil
	}
	return stream.subscribe(ctx)
}

func (s *streamer) Info(ctx context.Context) *LogStreamInfo {
	s.Lock()
	defer s.Unlock()
	info := &LogStreamInfo{
		Streams: map[int64]int{},
	}
	for id, stream := range s.streams {
		stream.Lock()
		info.Streams[id] = len(stream.list)
		stream.Unlock()
	}
	return info
}
