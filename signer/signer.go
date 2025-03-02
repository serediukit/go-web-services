package main

import (
	"fmt"
	"strconv"
)

//func ExecutePipeline(jobs ...job) {
//
//}

func SingleHash(in []int) []string {
	out := make([]string, len(in))
	for i, v := range in {
		data := strconv.Itoa(v)
		calc := DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
		fmt.Println("SH calc:", calc)
		out[i] = calc
	}
	return out
}

func MultiHash(in []string) []string {
	out := make([]string, len(in))
	for i, v := range in {
		str := ""
		for j := 0; j <= 5; j++ {
			calc := DataSignerCrc32(strconv.Itoa(j) + v)
			fmt.Println("MH calc:", calc)
			str += calc
		}
		out[i] = str
		fmt.Println("MH str:", str)
	}
	return out
}

func CombineResults(in []string) string {
	str := in[0]
	for i := 1; i < len(in); i++ {
		str += "_" + in[i]
	}
	fmt.Println("SR res:", str)
	return str
}

func main() {

	in := []int{0, 1}

	out1 := SingleHash(in)

	out2 := MultiHash(out1)

	out3 := CombineResults(out2)

	fmt.Println(out3)
}
