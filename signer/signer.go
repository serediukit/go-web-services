package main

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 100)

	wg := &sync.WaitGroup{}

	for i, j := range jobs {
		in = out
		out = make(chan interface{}, 100)

		fmt.Printf("JOB #%d\n", i)
		//fmt.Println("CHANNEL IN:", in)
		//fmt.Println("CHANNEL OUT:", out)

		wg.Add(1)
		go func(j job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)

			j(in, out)
		}(j, in, out)
	}

	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	for i := range in {
		data := strconv.Itoa(i.(int))
		md5 := DataSignerMd5(data)
		crc32Data := DataSignerCrc32(data)
		crc32Md5 := DataSignerCrc32(md5)
		calc := crc32Data + "~" + crc32Md5
		fmt.Println("SH calc:", calc)
		out <- calc
		runtime.Gosched()
	}
}

func MultiHash(in, out chan interface{}) {
	for i := range in {
		str := ""
		data := (i).(string)
		for j := 0; j <= 5; j++ {
			calc := DataSignerCrc32(strconv.Itoa(j) + data)
			fmt.Println("MH calc:", calc)
			str += calc
		}
		out <- str
		fmt.Println("MH str:", str)
	}
}

func CombineResults(in, out chan interface{}) {
	str := (<-in).(string)
	for i := range in {
		str += "_" + i.(string)
	}
	out <- str
}

func main() {

	jobStart := func(in, out chan interface{}) {
		out <- 0
		out <- 1
	}

	jobEnd := func(in, out chan interface{}) {
		for i := range in {
			fmt.Println("END GET:", i)
		}
	}

	ExecutePipeline(jobStart, SingleHash, MultiHash, CombineResults, jobEnd)
}
