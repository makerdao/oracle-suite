package query

import "fmt"

// max amount of tasks in worker pool queue
const maxTasksQueue = 10

// WorkerPool interface for any Query Engine worker pools
type WorkerPool interface {
	Start()
	Stop() error
	Query(req *HTTPRequest) *HTTPResponse
}

// HTTPWorkerPool structure that contain Woker Pool HTTP implementation
// It implements worker pool that will do real HTTP calls to resources using `query.MakeGetRequest`
type HTTPWorkerPool struct {
	started     bool
	workerCount int
	input       chan *asyncHTTPRequest
}

type asyncHTTPRequest struct {
	request  *HTTPRequest
	response chan *HTTPResponse
}

// NewHTTPWorkerPool create new worker pool for queries
func NewHTTPWorkerPool(workerCount int) *HTTPWorkerPool {
	return &HTTPWorkerPool{
		workerCount: workerCount,
		input:       make(chan *asyncHTTPRequest, maxTasksQueue),
	}
}

// Start start worker pool
func (wp *HTTPWorkerPool) Start() {
	for w := 1; w <= wp.workerCount; w++ {
		go worker(w, wp.input)
	}
	wp.started = true
}

// Stop stop worker in pool
func (wp *HTTPWorkerPool) Stop() error {
	close(wp.input)
	wp.started = false
	return nil
}

// Query makes request to given Request
// Under the hood it will wrap everything to async query and execute it using
// worker pool.
func (wp *HTTPWorkerPool) Query(req *HTTPRequest) *HTTPResponse {
	if !wp.started {
		return &HTTPResponse{
			Error: fmt.Errorf("worker pool not strated"),
		}
	}

	asyncReq := &asyncHTTPRequest{
		request:  req,
		response: make(chan *HTTPResponse),
	}
	// Sending request
	wp.input <- asyncReq
	// Waiting for response
	res := <-asyncReq.response
	// Have to close channel
	close(asyncReq.response)
	return res
}

func worker(id int, jobs <-chan *asyncHTTPRequest) {
	for req := range jobs {
		// Make request and return result into channel
		if req.response != nil {
			req.response <- MakeGetRequest(req.request)
		}
	}
}
