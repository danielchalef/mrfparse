/*
Copyright © 2023 Daniel Chalef

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
package utils

import (
	"math"
	"sort"
	"time"
)

// Time execution of a function
type wrapped func()

func Timed(fn wrapped) int64 {
	start := time.Now().Unix()

	fn()

	end := time.Now().Unix()

	return end - start
}

type TimedOperation func() time.Duration

// timeOperation performs the given operation and returns the duration of the operation
// in nanoseconds.
func timeOperation(op TimedOperation) time.Duration {
	start := time.Now()

	op()

	return time.Since(start)
}

// measureExecutionTimes takes a TimedOperation and a number of iterations as arguments
// and returns the mean, median, and variance of the execution times.
func MeasureExecutionTimes(op TimedOperation, iterations int) (mean, median, variance int64) {
	times := make([]int64, iterations)

	// Perform the operation and record the execution time for each iteration.
	for i := 0; i < iterations; i++ {
		times[i] = timeOperation(op).Nanoseconds()
	}

	// Calculate the mean execution time.
	var sum int64
	for _, t := range times {
		sum += t
	}

	meanF := float64(sum) / float64(iterations)

	// convert mean to an int64
	mean = int64(meanF)

	// Calculate the variance of the execution times.
	var varianceF float64
	for _, t := range times {
		varianceF += math.Pow(float64(t)-meanF, 2)
	}

	varianceF /= float64(iterations)
	variance = int64(varianceF)

	// Calculate the median execution time.
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	median = times[iterations/2]

	return mean, median, variance
}
