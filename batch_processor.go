package common

import (
	"log"
	"math"
	"sync"
	"time"
)

type ProcessFunction func([]interface{}, *sync.WaitGroup)

type BatchProcessor struct {
	ProcessFunc     ProcessFunction
	ProcessInterval time.Duration
	ThreadCount     int
	Events          chan interface{}
	Running         bool
	WaitGroup       sync.WaitGroup
}

func NewBatchProcessor(processFunc ProcessFunction, requestQueueSize int, processInterval int, threadCount int) *BatchProcessor {
	batchProcessor := new(BatchProcessor)

	batchProcessor.Events = make(chan interface{}, requestQueueSize)
	batchProcessor.ProcessInterval = time.Duration(processInterval)
	batchProcessor.ThreadCount = threadCount
	batchProcessor.ProcessFunc = processFunc

	return batchProcessor
}

func (batchProcessor *BatchProcessor) AddEvent(event interface{}) {
	batchProcessor.Events <- event
}

func (batchProcessor *BatchProcessor) Stop() {
	batchProcessor.Running = false
	batchProcessor.WaitGroup.Wait()
}

func (batchProcessor *BatchProcessor) Start() {
	go batchProcessor.process()
}

func (batchProcessor *BatchProcessor) process() {
	log.Print("Started batch writing thread")

	batchProcessor.Running = true
	batchProcessor.WaitGroup.Add(1)
	defer batchProcessor.WaitGroup.Done()

	for batchProcessor.Running {
		time.Sleep(batchProcessor.ProcessInterval * time.Second)

		var elements []interface{}
		processing := true
		for processing {
			select {
			case event, ok := <-batchProcessor.Events:
				if ok {
					elements = append(elements, event)
					break
				} else {
					log.Print("Select channel closed")
					processing = false
					batchProcessor.Running = false
					break
				}
			default:
				processing = false
				break
			}
		}

		if len(elements) <= 0 {
			continue
		}

		log.Printf("Retrieved %d values.  Processing with %d connections", len(elements), batchProcessor.ThreadCount)

		sliceSize := int(math.Floor(float64(len(elements) / batchProcessor.ThreadCount)))
		remainder := len(elements) % batchProcessor.ThreadCount
		start := 0
		end := 0

		for iter := 0; iter < batchProcessor.ThreadCount; iter++ {
			var leftover int
			if remainder > 0 {
				leftover = 1
				remainder--
			} else {
				leftover = 0
			}

			end += sliceSize + leftover

			if start == end {
				break
			}

			batchProcessor.WaitGroup.Add(1)
			go batchProcessor.ProcessFunc(elements[start:end], &batchProcessor.WaitGroup)

			start = end
		}
	}
}
