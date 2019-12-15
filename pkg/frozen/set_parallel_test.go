package frozen

import (
	"runtime"
	"sync"
	"testing"
)

func BenchmarkSetSequentialWith1M(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var s Set
		for i := 0; i < 1<<20; i++ {
			s = s.With(i)
		}
		if s.Count() != 1<<20 {
			b.Errorf("Wrong count: %x", s.Count())
		}
	}
}

func BenchmarkSetSequentialBuilder1M(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var sb SetBuilder
		for i := 0; i < 1<<20; i++ {
			sb.Add(i)
		}
		s := sb.Finish()
		if s.Count() != 1<<20 {
			b.Errorf("Wrong count: %x", s.Count())
		}
	}
}

func parallelUnion(sets []Set) Set {
	switch len(sets) {
	case 1:
		return sets[0]
	case 2:
		return sets[0].Union(sets[1])
	default:
		half := len(sets) / 2
		ch := make(chan Set)
		go func() {
			ch <- parallelUnion(sets[:half])
		}()
		return parallelUnion(sets[half:]).Union(<-ch)
	}
}

func BenchmarkSetParallelWith1M(b *testing.B) {
	D := runtime.GOMAXPROCS(0)
	for n := 0; n < b.N; n++ {
		sets := make([]Set, D)
		var wg sync.WaitGroup
		wg.Add(D)
		for d := 0; d < D; d++ {
			d := d
			go func() {
				s := &sets[d]
				for i := d; i < 1<<20; i += D {
					*s = s.With(i)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		s := parallelUnion(sets)
		if s.Count() != 1<<20 {
			b.Errorf("Wrong count: %x", s.Count())
		}
	}
}

func BenchmarkSetParallelBuilder1M(b *testing.B) {
	D := runtime.GOMAXPROCS(0)
	for n := 0; n < b.N; n++ {
		builders := make([]SetBuilder, D)
		var wg sync.WaitGroup
		wg.Add(D)
		for d := 0; d < D; d++ {
			d := d
			go func() {
				sb := &builders[d]
				for i := d; i < 1<<20; i += D {
					sb.Add(i)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		sets := make([]Set, 0, D)
		for _, builder := range builders {
			sets = append(sets, builder.Finish())
		}
		s := parallelUnion(sets)
		if s.Count() != 1<<20 {
			b.Errorf("Wrong count: %x", s.Count())
		}
	}
}
