package production

import (
	"fmt"
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/dmedovich/VRH/algorithms"
	"github.com/dmedovich/VRH/vrh"
	nspcc "github.com/nspcc-dev/hrw/v2"
)

var ownerSink uint64
var objectSink any

func ring(t testing.TB, nodes []uint64, backend vrh.Backend) *vrh.Ring {
	t.Helper()
	r, e := vrh.New(nodes, vrh.WithBackend(backend))
	if e != nil {
		t.Fatal(e)
	}
	return r
}

func TestProductionContract(t *testing.T) {
	for _, nodes := range [][]uint64{nil, {}, {1, 1}} {
		if _, err := vrh.New(nodes); err == nil {
			t.Fatal("invalid nodes accepted")
		}
	}
	for _, backend := range []vrh.Backend{vrh.BackendAVX2, 255} {
		if _, err := vrh.New([]uint64{1}, vrh.WithBackend(backend)); err == nil {
			t.Fatal("invalid backend accepted")
		}
	}
	if _, err := vrh.New([]uint64{1}, nil); err != nil {
		t.Fatal(err)
	}
	keys := Keys("contract", "random", 128)
	for _, v := range []uint64{0, 1, 1 << 63, ^uint64(0)} {
		keys = append(keys, NewKey(v))
	}
	sizes := make([]int, 257)
	for i := range sizes {
		sizes[i] = i + 1
	}
	sizes = append(sizes, 1024, 4096, 16384)
	for _, n := range sizes {
		nodes, _ := Dataset(n, "scattered")
		r := ring(t, nodes, vrh.BackendAuto)
		s := ring(t, nodes, vrh.BackendScalar)
		if r.Len() != n || r.Backend() != "scalar" || !slices.Equal(r.Nodes(), nodes) {
			t.Fatal("metadata")
		}
		rev := slices.Clone(nodes)
		slices.Reverse(rev)
		reversed := ring(t, rev, vrh.BackendAuto)
		for _, k := range keys {
			want := r.OwnerScalar(k.Value)
			if r.Owner(k.Value) != want || s.Owner(k.Value) != want || reversed.Owner(k.Value) != want {
				t.Fatalf("n=%d key=%x mapping mismatch", n, k.Value)
			}
		}
		raw := make([]uint64, len(keys))
		for i, k := range keys {
			raw[i] = k.Value
		}
		want := r.OwnerBatch(raw, nil)
		for _, dst := range [][]uint64{nil, make([]uint64, 1), make([]uint64, len(raw)), make([]uint64, len(raw)+2), slices.Clone(raw)} {
			if !slices.Equal(r.OwnerBatch(raw, dst), want) {
				t.Fatal("batch mismatch")
			}
		}
		alias := slices.Clone(raw)
		if !slices.Equal(r.OwnerBatch(alias, alias), want) {
			t.Fatal("in-place mismatch")
		}
		if len(r.OwnerBatch(nil, nil)) != 0 {
			t.Fatal("empty batch")
		}
		dst := make([]uint64, len(raw))
		if testing.AllocsPerRun(5, func() { r.OwnerBatch(raw, dst) }) != 0 {
			t.Fatal("batch alloc")
		}
		if testing.AllocsPerRun(5, func() { ownerSink = r.Owner(42) }) != 0 {
			t.Fatal("owner alloc")
		}
		copyNodes := r.Nodes()
		copyNodes[0] ^= 1
		nodes[0] ^= 1
		if r.Nodes()[0] != s.Nodes()[0] {
			t.Fatal("mutation escaped")
		}
	}
}

func TestMembershipAndConcurrentReaders(t *testing.T) {
	for _, n := range []int{3, 16, 17, 64, 256} {
		nodes, _ := Dataset(n, "scattered")
		r := ring(t, nodes, vrh.BackendAuto)
		const added = 0xfedcba9876543210
		expanded := ring(t, append(slices.Clone(nodes), added), vrh.BackendAuto)
		reduced := ring(t, nodes[1:], vrh.BackendAuto)
		keys := Keys("membership", "random", 4096)
		for _, key := range keys {
			before := r.Owner(key.Value)
			if after := expanded.Owner(key.Value); after != before && after != added {
				t.Fatal("addition")
			}
			if after := reduced.Owner(key.Value); before != nodes[0] && after != before {
				t.Fatal("removal")
			}
		}
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				for _, key := range keys {
					if r.Owner(key.Value) != r.OwnerScalar(key.Value) {
						t.Error("concurrent mismatch")
						return
					}
				}
			})
		}
		wg.Wait()
	}
}

func TestNSPCCAdapter(t *testing.T) {
	for _, n := range []int{1, 2, 3, 16, 17, 64, 257} {
		nodes, _ := Dataset(n, "scattered")
		scan := NewNSPCCScan(nodes)
		hashed := PrepareNSPCC(nodes)
		for _, key := range Keys("nspcc", "random", 128) {
			work := slices.Clone(hashed)
			nspcc.Sort(work, nspcc.WrapBytes(key.Bytes[:]))
			if scan.Owner(key) != work[0].ID {
				t.Fatal("NSPCC scan differs from upstream Sort")
			}
		}
	}
}

func BenchmarkBuild(b *testing.B) {
	for _, n := range Sizes {
		nodes, _ := Dataset(n, "scattered")
		for _, name := range Names {
			b.Run(fmt.Sprintf("n=%d/%s", n, name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					objectSink = Build(name, nodes)
				}
			})
		}
	}
}

// DirectNew excludes the common comparator closure and interface wrapper.
func BenchmarkDirectNew(b *testing.B) {
	for _, n := range Sizes {
		nodes, _ := Dataset(n, "scattered")
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				r, err := vrh.New(nodes)
				if err != nil {
					b.Fatal(err)
				}
				objectSink = r
			}
		})
	}
}

func BenchmarkBackend(b *testing.B) {
	for _, n := range []int{3, 16, 17, 64, 1024, 16384} {
		nodes, _ := Dataset(n, "scattered")
		keys := Keys("backend", "random", 65536)
		for _, mode := range []string{"auto", "scalar", "reference"} {
			r := ring(b, nodes, vrh.BackendAuto)
			if mode == "scalar" {
				r = ring(b, nodes, vrh.BackendScalar)
			}
			owner := r.Owner
			if mode == "reference" {
				owner = r.OwnerScalar
			}
			b.Run(fmt.Sprintf("n=%d/%s", n, mode), func(b *testing.B) {
				b.ReportAllocs()
				var v uint64
				for i := 0; b.Loop(); i++ {
					v ^= owner(keys[i&65535].Value)
				}
				ownerSink = v
			})
		}
	}
}

func BenchmarkBatch(b *testing.B) {
	for _, n := range []int{3, 16, 17, 256, 16384} {
		nodes, _ := Dataset(n, "scattered")
		r := ring(b, nodes, vrh.BackendAuto)
		for _, length := range []int{0, 1, 16, 64, 1024} {
			keys := Keys("batch", "random", length)
			raw := make([]uint64, length)
			for i, k := range keys {
				raw[i] = k.Value
			}
			for _, mode := range []string{"reuse", "nil", "short", "inplace"} {
				b.Run(fmt.Sprintf("n=%d/k=%d/vrh-ring/%s", n, length, mode), func(b *testing.B) {
					dst := make([]uint64, length)
					alias := slices.Clone(raw)
					var result []uint64
					b.ReportAllocs()
					for b.Loop() {
						switch mode {
						case "reuse":
							result = r.OwnerBatch(raw, dst)
						case "nil":
							result = r.OwnerBatch(raw, nil)
						case "short":
							result = r.OwnerBatch(raw, dst[:max(0, length-1)])
						case "inplace":
							result = r.OwnerBatch(alias, alias)
						}
					}
					objectSink = result
				})
			}
			for _, impl := range ExternalImplementations(nodes) {
				for _, mode := range []string{"reuse", "nil"} {
					b.Run(fmt.Sprintf("n=%d/k=%d/%s/%s", n, length, impl.Name, mode), func(b *testing.B) {
						dst := make([]uint64, length)
						b.ReportAllocs()
						for b.Loop() {
							if mode == "nil" {
								dst = make([]uint64, length)
							}
							for i, k := range keys {
								dst[i] = impl.Owner(k)
							}
						}
						objectSink = dst
					})
				}
			}
		}
	}
}

func BenchmarkRebuild(b *testing.B) {
	for _, n := range []int{3, 17, 256, 16384} {
		nodes, _ := Dataset(n, "scattered")
		keys := Keys("rebuild", "random", 4096)
		for _, q := range []int{1, 64, 4096} {
			for _, name := range Names {
				b.Run(fmt.Sprintf("n=%d/q=%d/%s", n, q, name), func(b *testing.B) {
					work := slices.Clone(nodes)
					var v uint64
					b.ReportAllocs()
					for i := 0; b.Loop(); i++ {
						work[n-1] = ^uint64(i & 1)
						impl := Build(name, work)
						for _, k := range keys[:q] {
							v ^= impl.Owner(k)
						}
					}
					ownerSink = v
				})
			}
		}
	}
}

func BenchmarkMetadata(b *testing.B) {
	for _, n := range []int{3, 16, 17, 256, 16384} {
		nodes, _ := Dataset(n, "scattered")
		r := ring(b, nodes, vrh.BackendAuto)
		b.Run(fmt.Sprintf("n=%d/Nodes", n), func(b *testing.B) {
			var dst []uint64
			b.ReportAllocs()
			for b.Loop() {
				dst = r.Nodes()
			}
			objectSink = dst
		})
		b.Run(fmt.Sprintf("n=%d/Len", n), func(b *testing.B) {
			var v int
			for b.Loop() {
				v = r.Len()
			}
			ownerSink = uint64(v)
		})
		b.Run(fmt.Sprintf("n=%d/Backend", n), func(b *testing.B) {
			var v string
			for b.Loop() {
				v = r.Backend()
			}
			objectSink = v
		})
	}
}

func BenchmarkParallel(b *testing.B) {
	for _, n := range []int{16, 256, 16384} {
		nodes, _ := Dataset(n, "scattered")
		keys := Keys("parallel", "random", 65536)
		for _, impl := range Implementations(nodes) {
			b.Run(fmt.Sprintf("n=%d/%s", n, impl.Name), func(b *testing.B) {
				b.ReportAllocs()
				b.RunParallel(func(pb *testing.PB) {
					var v uint64
					i := 0
					for pb.Next() {
						v ^= impl.Owner(keys[i&65535])
						i++
					}
					runtime.KeepAlive(v)
				})
			})
		}
	}
}

// Sanity-check deliberately difficult keys without assuming distinct outputs.
func TestKeyBoundaries(t *testing.T) {
	nodes := []uint64{0, 1, 1 << 63, ^uint64(0)}
	r := ring(t, nodes, vrh.BackendAuto)
	for _, key := range nodes {
		scores := map[uint64]bool{}
		for _, n := range nodes {
			s := algorithms.VRHScore(key, n)
			if scores[s] {
				t.Fatal("tie")
			}
			scores[s] = true
		}
		if !slices.Contains(nodes, r.Owner(key)) {
			t.Fatal("nonmember")
		}
	}
}
