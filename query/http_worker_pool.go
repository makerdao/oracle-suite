//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package query

import "fmt"

// max amount of tasks in worker pool queue
const maxTasksQueue = 10

// WorkerPool interface for any Query Engine worker pools
type WorkerPool interface {
	Ready() bool
	Start()
	Stop() error
	Query(req *HTTPRequest) *HTTPResponse
}

// HTTPWorkerPool structure that contain Woker Pool HTTP implementation
// It implements worker pool that will do real HTTP calls to resources using `query.MakeHTTPRequest`
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

// Ready checks if pool is ready to be used
func (wp *HTTPWorkerPool) Ready() bool {
	return wp.started
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
			req.response <- MakeHTTPRequest(req.request)
		}
	}
}
