package query

import "fmt"

// max amount of tasks in worker pool queue
const maxTasksQueue = 10

// WorkerPool structure that contain Woker Pool implementation
type WorkerPool struct {
	workerCount int
	input       chan *asyncHTTPRequest
}

type asyncHTTPRequest struct {
	request  *HTTPRequest
	response chan *HTTPResponse
}

// NewWorkerPool create new worker pool for queries
func NewWorkerPool(workerCount int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		input:       make(chan *asyncHTTPRequest, maxTasksQueue),
	}
}

// Start start worker pool
func (wp *WorkerPool) Start() {
	for w := 1; w <= wp.workerCount; w++ {
		go worker(w, wp.input)
	}
}

// Stop stop worker in pool
func (wp *WorkerPool) Stop() error {
	close(wp.input)
	return nil
}

// Query asdf
func (wp *WorkerPool) Query(req *HTTPRequest) *HTTPResponse {
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
		fmt.Println("worker", id, "making request")
		// Make request and return result into channel
		if req.response != nil {
			req.response <- MakeGetRequest(req.request)
		}
	}
}
