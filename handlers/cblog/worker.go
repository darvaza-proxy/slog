package cblog

// NewWithCallback creates a logger and attached a goroutine to call
// a provided handler, sequentially, for each log entry
func NewWithCallback(size int, h func(LogMsg)) *Logger {
	// TODO: default handler?
	if h != nil {
		if size <= 0 {
			size = DefaultOutputBufferSize
		}

		return newWithWorker(size, h)
	}
	return nil
}

func newWithWorker(size int, h func(LogMsg)) *Logger {
	ch := make(chan LogMsg, size)
	l := newLogger(ch)

	go func() {
		defer close(ch)

		for msg := range ch {
			h(msg)
		}
	}()

	return &l.Logger
}
