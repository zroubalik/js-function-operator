/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/knative/serving/test"
)

// Algorithm from https://stackoverflow.com/a/21854246

// Only primes less than or equal to N will be generated
func primes(N int) []int {

	var x, y, n int
	nsqrt := math.Sqrt(float64(N))

	isPrime := make([]bool, N)

	for x = 1; float64(x) <= nsqrt; x++ {
		for y = 1; float64(y) <= nsqrt; y++ {
			n = 4*(x*x) + y*y
			if n <= N && (n%12 == 1 || n%12 == 5) {
				isPrime[n] = !isPrime[n]
			}
			n = 3*(x*x) + y*y
			if n <= N && n%12 == 7 {
				isPrime[n] = !isPrime[n]
			}
			n = 3*(x*x) - y*y
			if x > y && n <= N && n%12 == 11 {
				isPrime[n] = !isPrime[n]
			}
		}
	}

	for n = 5; float64(n) <= nsqrt; n++ {
		if isPrime[n] {
			for y = n * n; y < N; y += n * n {
				isPrime[y] = false
			}
		}
	}

	isPrime[2] = true
	isPrime[3] = true

	primes := make([]int, 0, 1270606)
	for x = 0; x < len(isPrime)-1; x++ {
		if isPrime[x] {
			primes = append(primes, x)
		}
	}

	// primes is now a slice that contains all primes numbers up to N
	return primes
}

func bloat(mb int) string {
	b := make([]byte, mb*1024*1024)
	for i := 0; i < len(b); i++ {
		b[i] = 1
	}
	return fmt.Sprintf("Allocated %v Mb of memory.\n", mb)
}

func prime(max int) string {
	p := primes(max)
	if len(p) == 0 {
		return fmt.Sprintf("There are no primes smaller than %d.\n", max)
	}
	return fmt.Sprintf("The largest prime less than %d is %d.\n", max, p[len(p)-1])
}

func sleep(ms int) string {
	start := time.Now()
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return fmt.Sprintf("Slept for %v.\n", time.Since(start))
}

func parseIntParam(r *http.Request, param string) (int, bool, error) {
	value := r.URL.Query().Get(param)
	if value == "" {
		return 0, false, nil
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, false, err
	}
	return i, true, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Validate inputs.
	ms, hasMs, err := parseIntParam(r, "sleep")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	max, hasMax, err := parseIntParam(r, "prime")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mb, hasMb, err := parseIntParam(r, "bloat")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Consume time, cpu and memory in parallel.
	var wg sync.WaitGroup
	defer wg.Wait()
	if hasMs && ms > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Fprint(w, sleep(ms))
		}()
	}
	if hasMax && max > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Fprint(w, prime(max))
		}()
	}
	if hasMb && mb > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Fprint(w, bloat(mb))
		}()
	}
}

func main() {
	test.ListenAndServeGracefully(":8080", handler)
}
