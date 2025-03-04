package main

import (
	"fmt"
	"runtime"
	"strconv"
)

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 1)
	out := make(chan interface{}, 2)

	for i, job := range jobs {
		fmt.Printf("JOB #%d\n", i)
		//fmt.Println("CHANNEL IN:", in)
		//fmt.Println("CHANNEL OUT:", out)
		go job(in, out)
		in = out
		out = make(chan interface{}, 2)
	}
}

func SingleHash(in, out chan interface{}) {
	data := strconv.Itoa((<-in).(int))
	calc := DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
	fmt.Println("SH calc:", calc)
	out <- calc
	runtime.Gosched()
}

func MultiHash(in, out chan interface{}) {
	str := ""
	data := (<-in).(string)
	for j := 0; j <= 5; j++ {
		calc := DataSignerCrc32(strconv.Itoa(j) + data)
		fmt.Println("MH calc:", calc)
		str += calc
	}
	out <- str
	fmt.Println("MH str:", str)

}

//func CombineResults(in []string) string {
//	str := in[0]
//	for i := 1; i < len(in); i++ {
//		str += "_" + in[i]
//	}
//	fmt.Println("SR res:", str)
//	return str
//}

func main() {

	jobStart := func(in, out chan interface{}) {
		out <- 0
		out <- 1
	}

	jobEnd := func(in, out chan interface{}) {
		fmt.Println("END GET:", <-in)
	}

	ExecutePipeline(jobStart, SingleHash, MultiHash, jobEnd)

	fmt.Scanln()
}
