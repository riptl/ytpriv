package worker

import "context"

type workerContext struct{
	// Worker management
	ctxt context.Context

	// Sent when a worker gets
	// an error
	errors chan error

	// Job queue
	jobs chan string

	// Sent when a worker
	// processed a video
	results chan string

	// Sent whenever the
	// queue goes idle
	idle chan bool
}
