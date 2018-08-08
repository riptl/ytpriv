package worker

import "context"

type workerContext struct{
	// Worker management
	ctxt context.Context

	// Sent when a worker gets
	// an error
	errors chan error

	// Sent when a worker exists
	// because it's idle
	idleExists chan bool

	// Sent when a worker
	// processes a video
	results chan string
}
