package main

import (
	"fmt"
	"strconv"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 100)

	wg := &sync.WaitGroup{}

	for _, jb := range jobs {
		in = out
		out = make(chan interface{}, 100)

		wg.Add(1)
		go func(jb job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)

			jb(in, out)
		}(jb, in, out)
	}

	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for i := range in {
		wg.Add(1)
		data := strconv.Itoa(i.(int))

		go func(data string) {
			defer wg.Done()

			crc32DataCh := make(chan string)
			go func() {
				defer close(crc32DataCh)

				crc32DataCh <- DataSignerCrc32(data)
			}()

			crc32Md5Ch := make(chan string)
			go func() {
				defer close(crc32Md5Ch)

				mu.Lock()
				md5 := DataSignerMd5(data)
				mu.Unlock()

				crc32Md5Ch <- DataSignerCrc32(md5)
			}()

			calc := <-crc32DataCh + "~" + <-crc32Md5Ch

			out <- calc
		}(data)
	}

	wg.Wait()
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
		out <- 2
	}

	jobEnd := func(in, out chan interface{}) {
		for i := range in {
			fmt.Println("END GET:", i)
		}
	}

	ExecutePipeline(jobStart, SingleHash, MultiHash, CombineResults, jobEnd)
}
