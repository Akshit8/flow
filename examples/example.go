package main

import (
	"fmt"
	"time"

	"github.com/Akshit8/flow"
)

func add1(input interface{}) interface{} {
	if val, ok := input.(int); ok {
		return val + 1
	}

	return nil
}

func slowSquare(input interface{}) interface{} {
	time.Sleep(time.Second)

	if val, ok := input.(int); ok {
		return val * val
	}

	return nil
}

func slowPrint(input interface{}) interface{} {
	time.Sleep(100 * time.Millisecond)

	fmt.Println(input)

	return nil
}

func main() {
	input := make(chan interface{}, 100)
	for i := 0; i < 10; i++ {
		input <- i
	}
	close(input)

	p := flow.NewPipeline()

	p.AddSegmentWithCapacity(slowSquare, 6)
	p.AddSegment(add1)
	p.AddSegmentWithCapacity(slowPrint, 10)

	<-p.Run(input)
}
