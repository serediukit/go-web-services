package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

const multiHashThreadsNum = 6

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

			out <- <-crc32DataCh + "~" + <-crc32Md5Ch
		}(data)
	}

	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for i := range in {
		wg.Add(1)
		data := i.(string)

		go func(data string) {
			defer wg.Done()

			threadWG := &sync.WaitGroup{}
			threadRes := make([]string, multiHashThreadsNum)

			for th := range multiHashThreadsNum {
				threadWG.Add(1)

				go func(thIndex int) {
					defer threadWG.Done()

					val := strconv.Itoa(th) + data

					threadRes[thIndex] = DataSignerCrc32(val)
				}(th)
			}

			threadWG.Wait()

			out <- strings.Join(threadRes, "")
		}(data)
	}

	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var res []string
	for i := range in {
		res = append(res, i.(string))
	}
	sort.Strings(res)
	out <- strings.Join(res, "_")
}
