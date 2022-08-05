package flow

import "sync"

type Pump func(inputObject interface{}) (outputObject interface{})

type Segment func(input <-chan interface{}) (output chan interface{})

type Pipeline []Segment

func NewPipeline() Pipeline {
	return make(Pipeline, 0)
}

func (p *Pipeline) AddSegment(pump Pump) {
	*p = append(*p, constructSegment(pump, 1))
}

func (p *Pipeline) AddSegmentWithCapacity(pump Pump, capacity int) {
	*p = append(*p, constructSegment(pump, capacity))
}

func (p *Pipeline) Run(input <-chan interface{}) chan struct{} {
	for _, segment := range *p {
		// reassign input to the output of the previous segment
		// this way when next segment is called, it will receive the output of the previous segment
		input = segment(input)
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		for range input {
		}
	}()

	return done
}

func embedPumpInSegment(pump Pump) Segment {
	return func(input <-chan interface{}) chan interface{} {
		output := make(chan interface{})
		go func() {
			defer close(output)
			for inputObject := range input {
				outputObject := pump(inputObject)
				if outputObject != nil {
					output <- outputObject
				}
			}
		}()

		return output
	}
}

func constructSegment(pump Pump, capacity int) Segment {
	return func(input <-chan interface{}) chan interface{} {
		var channels []chan interface{}

		for i := 0; i < capacity; i++ {
			channels = append(channels, embedPumpInSegment(pump)(input))
		}

		return MergeChannels(channels)
	}
}

func MergeChannels(channels []chan interface{}) chan interface{} {
	var wg sync.WaitGroup

	wg.Add(len(channels))

	output := make(chan interface{})

	for _, channel := range channels {
		go func(channel chan interface{}) {
			defer wg.Done()
			for inputObject := range channel {
				output <- inputObject
			}
		}(channel)
	}

	go func() {
		defer close(output)
		wg.Wait()
	}()

	return output
}
