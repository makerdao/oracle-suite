package query

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
}

// Stop stop worker in pool
func (wp *HTTPWorkerPool) Stop() error {
	close(wp.input)
	return nil
}

// Query asdf
func (wp *HTTPWorkerPool) Query(req *HTTPRequest) *HTTPResponse {
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
