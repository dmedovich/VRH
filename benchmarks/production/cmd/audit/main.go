package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"slices"
	"time"

	p "github.com/dmedovich/VRH/benchmarks/production"
)

type row struct {
	Case           string    `json:"case"`
	Implementation string    `json:"implementation"`
	Nodes          int       `json:"nodes"`
	Samples        int       `json:"samples"`
	Stream         string    `json:"stream"`
	Counts         []uint64  `json:"counts"`
	Runs           []float64 `json:"runs_ns"`
	Median         float64   `json:"median_ns"`
	MaxPP          float64   `json:"max_pp"`
	MaxZ           float64   `json:"max_z"`
	Chi            float64   `json:"chi_square"`
	CVRatio        float64   `json:"cv_over_noise"`
	Pass           bool      `json:"pass"`
}

var sink uint64

func main() {
	mode := flag.String("mode", "matrix", "matrix, holdout, smoke")
	flag.Parse()
	runtime.GOMAXPROCS(1)
	enc := json.NewEncoder(os.Stdout)
	run := func(label string, nodes []uint64, stream string, samples, reps int) {
		keys := p.Keys(label, stream, samples)
		impls := p.Implementations(nodes)
		index := make(map[uint64]int, len(nodes))
		for i, n := range nodes {
			index[n] = i
		}
		rows := make([]row, len(impls))
		for i, impl := range impls {
			r := row{Case: label, Implementation: impl.Name, Nodes: len(nodes), Samples: samples, Stream: stream, Counts: make([]uint64, len(nodes))}
			for _, key := range keys {
				owner := impl.Owner(key)
				j, ok := index[owner]
				if !ok {
					panic("nonmember")
				}
				r.Counts[j]++
			}
			mean := float64(samples) / float64(len(nodes))
			sigma := math.Sqrt(mean * (1 - 1/float64(len(nodes))))
			for _, count := range r.Counts {
				d := float64(count) - mean
				r.MaxPP = math.Max(r.MaxPP, math.Abs(d)/float64(samples)*100)
				r.Chi += d * d / mean
				if sigma > 0 {
					r.MaxZ = math.Max(r.MaxZ, math.Abs(d)/sigma)
				}
			}
			if len(nodes) > 1 {
				r.CVRatio = math.Sqrt(r.Chi / float64(len(nodes)-1))
			}
			r.Pass = r.MaxZ <= 7 && r.Chi <= float64(len(nodes)-1)+8*math.Sqrt(2*float64(len(nodes)-1))
			rows[i] = r
		}
		runtime.GC()
		oldGC := debug.SetGCPercent(-1)
		for rep := 0; rep < reps; rep++ {
			for pos := range impls {
				i := (rep + pos) % len(impls)
				impl := impls[i]
				for _, key := range keys[:min(4096, len(keys))] {
					sink ^= impl.Owner(key)
				}
				var sum uint64
				start := time.Now()
				for _, key := range keys {
					sum ^= impl.Owner(key)
				}
				elapsed := time.Since(start)
				sink ^= sum
				rows[i].Runs = append(rows[i].Runs, float64(elapsed.Nanoseconds())/float64(samples))
			}
		}
		debug.SetGCPercent(oldGC)
		for i := range rows {
			r := &rows[i]
			if len(r.Runs) > 0 {
				v := slices.Clone(r.Runs)
				slices.Sort(v)
				r.Median = v[len(v)/2]
			}
			if err := enc.Encode(r); err != nil {
				panic(err)
			}
		}
		fmt.Fprintln(os.Stderr, "completed", label, stream, samples)
	}
	if *mode == "holdout" {
		for _, set := range p.KnownSets() {
			run("holdout-"+set.Name, set.Nodes, "random", 1<<20, 0)
		}
		nodes, _ := p.Dataset(16384, "scattered")
		run("holdout-large", nodes, "random", 1<<20, 0)
		return
	}
	if *mode == "smoke" {
		nodes, _ := p.Dataset(17, "scattered")
		run("smoke", nodes, "random", 4096, 1)
		return
	}
	if *mode != "matrix" {
		panic("unknown mode")
	}
	for _, layout := range []string{"sequential", "scattered"} {
		for _, n := range p.Sizes {
			nodes, _ := p.Dataset(n, layout)
			samples := 1 << 18
			if n > 64 {
				samples = 1 << 16
			}
			if n > 1024 {
				samples = 1 << 14
			}
			for _, stream := range []string{"random", "sequential", "highbits"} {
				run(fmt.Sprintf("%s-%d", layout, n), nodes, stream, samples, 5)
			}
		}
	}
	for _, set := range p.KnownSets() {
		for _, stream := range []string{"random", "sequential", "highbits"} {
			run("known-"+set.Name, set.Nodes, stream, 1<<18, 5)
		}
	}
}
