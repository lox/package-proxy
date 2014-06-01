package ubuntu

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"sort"
	"time"
)

const (
	mirrorsUrl       = "http://mirrors.ubuntu.com/mirrors.txt"
	benchmarkUrl     = "dists/saucy/main/binary-amd64/Packages.bz2"
	benchmarkTimes   = 2
	benchmarkBytes   = 1024 * 256 // 256Kb
	benchmarkTimeout = 20         // 20 seconds
)

type Mirrors struct {
	URLs []string
}

func GetGeoMirrors() (m Mirrors, err error) {
	response, err := http.Get(mirrorsUrl)
	if err != nil {
		return
	}

	defer response.Body.Close()
	scanner := bufio.NewScanner(response.Body)
	m.URLs = []string{}

	// read urls line by line
	for scanner.Scan() {
		m.URLs = append(m.URLs, scanner.Text())
	}

	return m, scanner.Err()
}

func (m Mirrors) Fastest() (string, error) {
	ch := make(chan benchmarkResult)

	for _, url := range m.URLs {
		go func(u string) {
			duration, err := m.benchmark(u, benchmarkTimes)
			if err == nil {
				ch <- benchmarkResult{u, duration}
			}
		}(url)
	}

	results, err := m.readResults(ch, len(m.URLs))
	if len(results) == 0 {
		return "", errors.New("No results found: " + err.Error())
	} else if err != nil {
		log.Printf("Error benchmarking mirrors: %s", err.Error())
	}

	sort.Sort(results)
	return results[0].URL, nil
}

func (m Mirrors) readResults(ch <-chan benchmarkResult, size int) (br benchmarkResults, err error) {
	for {
		select {
		case r := <-ch:
			br = append(br, r)
			if len(br) == size {
				return
			}
		case <-time.After(benchmarkTimeout * time.Second):
			return br, errors.New("Timed out waiting for results")
		}
	}
}

func (m Mirrors) benchmark(url string, times int) (time.Duration, error) {
	var sum int64
	var d time.Duration

	url = url + benchmarkUrl

	for i := 0; i < times; i++ {
		timer := time.Now()
		response, err := http.Get(url)
		if err != nil {
			return d, err
		}

		defer response.Body.Close()
		_, err = io.ReadAtLeast(response.Body, make([]byte, benchmarkBytes), benchmarkBytes)
		if err != nil {
			return d, err
		}

		sum = sum + int64(time.Since(timer))
	}

	return time.Duration(sum / int64(times)), nil
}

type benchmarkResult struct {
	URL      string
	Duration time.Duration
}

type benchmarkResults []benchmarkResult

func (b benchmarkResults) Len() int {
	return len(b)
}

func (b benchmarkResults) Less(i, j int) bool {
	return b[i].Duration < b[j].Duration
}

func (b benchmarkResults) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
