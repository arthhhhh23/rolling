package rolling

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func ExampleNewWindow() {
	w := NewWindow(6, time.Second)
	w.Add(1)
	w.Add(2)
	time.Sleep(time.Second)
	w.Add(3)
	w.Add(4)
	w.Add(5)
	w.Add(6)
	w.Add(7)
	fmt.Println(w.Min(), w.Avg(), w.Max())
	// Output: 3 5 7
}

const tolerance = 0.00000000001

func TestMin(t *testing.T) {
	tests := []struct {
		name   string
		wsize  int
		values []float64
		expect float64
	}{
		{
			"zero values",
			3,
			nil,
			math.NaN(),
		},
		{
			"1 value",
			3,
			[]float64{42},
			float64(42),
		},
		{
			"first in window",
			3,
			[]float64{1, 2, 3},
			float64(1),
		},
		{
			"middle in window",
			3,
			[]float64{2, 1, 3},
			float64(1),
		},
		{
			"last in window",
			3,
			[]float64{2, 3, 1},
			float64(1),
		},
		{
			"last in window, evict",
			3,
			[]float64{1, 3, 4, 2},
			float64(2),
		},
		{
			"first in window, evict same",
			3,
			[]float64{1, 1, 4, 2},
			float64(1),
		},
		{
			"first in window, evict",
			3,
			[]float64{1, 2, 4, 3},
			float64(2),
		},
		{
			"middle in window, evict",
			3,
			[]float64{1, 3, 2, 5},
			float64(2),
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			w := NewWindow(int64(tcase.wsize), time.Hour)
			for _, v := range tcase.values {
				w.Add(v)
			}
			got := w.Min()
			if math.Abs(got-tcase.expect) > tolerance {
				t.Errorf("expected min %f, but got %f", tcase.expect, got)
			}
		})
	}
}

func TestEvictAll(t *testing.T) {
	w := NewWindow(3, time.Millisecond*5)
	w.Add(1)
	w.Add(2)
	w.Add(3)
	w.Add(4)
	w.Add(5)
	w.Add(6)
	w.Add(7)
	time.Sleep(time.Millisecond * 100)
	w.Evict()

	if w.Count() != 0 {
		t.Errorf("expected count 0, but got %d", w.Count())
	}

	if !math.IsNaN(w.First()) {
		t.Errorf("expected first NaN, but got %f", w.First())
	}

	if !math.IsNaN(w.Last()) {
		t.Errorf("expected last NaN, but got %f", w.Last())
	}

	if !math.IsNaN(w.Mid()) {
		t.Errorf("expected mid NaN, but got %f", w.Mid())
	}

	if !math.IsNaN(w.Avg()) {
		t.Errorf("expected avg NaN, but got %f", w.Avg())
	}

	if !math.IsNaN(w.Max()) {
		t.Errorf("expected max NaN, but got %f", w.Max())
	}

	if !math.IsNaN(w.Min()) {
		t.Errorf("expected min NaN, but got %f", w.Min())
	}
}

func TestAvg(t *testing.T) {
	tests := []struct {
		name   string
		wsize  int
		values []float64
		expect float64
	}{
		{
			"zero values",
			3,
			nil,
			math.NaN(),
		},
		{
			"1 value",
			3,
			[]float64{42},
			float64(42),
		},
		{
			"3 values",
			3,
			[]float64{1, 2, 3},
			float64(2),
		},
		{
			"5 evict",
			3,
			[]float64{1, 3, 4, 2, 3},
			float64(3),
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			w := NewWindow(int64(tcase.wsize), time.Hour)
			for _, v := range tcase.values {
				w.Add(v)
			}
			got := w.Avg()
			if math.Abs(got-tcase.expect) > tolerance {
				t.Errorf("expected avg %f, but got %f", tcase.expect, got)
			}
		})
	}
}

func TestOther(t *testing.T) {
	w := NewWindow(3, time.Second)
	for i := 0; i < 1000; i++ {
		w.Add(rand.Float64())
	}
	time.Sleep(time.Second)

	w.Add(1)
	w.Add(2)
	w.Add(3)

	if w.Count() != 3 {
		t.Errorf("expected count 3, but got %d", w.Count())
	}

	if math.Abs(w.First()-1) > tolerance {
		t.Errorf("expected first 1, but got %f", w.First())
	}

	if math.Abs(w.Last()-3) > tolerance {
		t.Errorf("expected last 3, but got %f", w.Last())
	}

	if math.Abs(w.Mid()-2) > tolerance {
		t.Errorf("expected mid 2, but got %f", w.Mid())
	}

	if math.Abs(w.Avg()-2) > tolerance {
		t.Errorf("expected avg 2, but got %f", w.Avg())
	}

	if math.Abs(w.Max()-3) > tolerance {
		t.Errorf("expected max 3, but got %f", w.Max())
	}

	if math.Abs(w.Min()-1) > tolerance {
		t.Errorf("expected min 1, but got %f", w.Min())
	}

	if math.Abs(w.Sum()-6) > tolerance {
		t.Errorf("expected sum 6, but got %f", w.Sum())
	}
}

func arrayMax(a []float64) float64 {
	max := -math.MaxFloat64
	for _, v := range a {
		if v > max {
			max = v
		}
	}
	return max
}

func arrayMin(a []float64) float64 {
	min := math.MaxFloat64
	for _, v := range a {
		if v < min {
			min = v
		}
	}
	return min
}

func arrayAvg(a []float64) float64 {
	sum := float64(0)
	for _, v := range a {
		sum += v
	}
	return sum / float64(len(a))
}

func TestFuzz(t *testing.T) {
	const iterations = 1000
	for sz := int64(1); sz < 1000; sz++ {
		w := NewWindow(sz, time.Minute)
		window := make([]float64, sz)
		for i := int64(0); i < iterations; i++ {
			value := rand.Float64()
			w.Add(value)
			window[i%sz] = value
			if i > sz {
				avg := w.Avg()
				exp := arrayAvg(window)
				if math.Abs(avg-exp) > tolerance {
					t.Fatalf("%d: expected avg %f, got %f", i, exp, avg)
				}
				min := w.Min()
				exp = arrayMin(window)
				if math.Abs(min-exp) > tolerance {
					t.Fatalf("%d: expected min %f, got %f", i, exp, min)
				}
				max := w.Max()
				exp = arrayMax(window)
				if math.Abs(max-exp) > tolerance {
					t.Fatalf("%d: expected max %f, got %f", i, exp, max)
				}
			}
		}
	}
}

func TestEvictAllAndAddAgain(t *testing.T) {
	w := NewWindow(3, time.Millisecond*100)
	w.AddAt(rand.Float64(), time.Now().Add(-time.Hour))
	w.Evict()
	w.Add(10.0)
	if w.Sum() != 10.0 {
		t.Errorf("expected sum 10, but got %f", w.Sum())
	}
}

func TestEvictAddAtOutOfOrder(t *testing.T) {
	w := NewWindow(100, time.Millisecond*100)
	before := runtime.MemStats{}
	runtime.GC()
	runtime.ReadMemStats(&before)
	for i := 0; i < 10000; i++ {
		w.AddAt(rand.Float64(), time.Now().Add((50*time.Millisecond)-time.Duration(rand.Intn(100))*time.Millisecond))
	}
	time.Sleep(time.Millisecond * 201)
	w.Evict()
	if w.Count() != 0 {
		t.Errorf("expected count 0, but got %d", w.Count())
	}
	if w.head != nil {
		t.Errorf("expected head nil, but got %v", w.head)
	}
	if w.tail != nil {
		t.Errorf("expected tail nil, but got %v", w.tail)
	}
	if w.maxDeque.Len() != 0 {
		t.Errorf("expected maxDeque len 0, but got %d", w.maxDeque.Len())
	}
	if w.minDeque.Len() != 0 {
		t.Errorf("expected minDeque len 0, but got %d", w.minDeque.Len())
	}
	after := runtime.MemStats{}
	runtime.GC()
	runtime.ReadMemStats(&after)
	if after.HeapAlloc > before.HeapAlloc {
		t.Errorf("expected heap alloc difference to be zero, but got %d bytes", after.HeapAlloc-before.HeapAlloc)
	}
}

func BenchmarkWindow_Add_10k(b *testing.B) {
	w := NewWindow(10_000, time.Second)
	for i := 0; i < b.N; i++ {
		w.Add(rand.Float64())
	}
}

func BenchmarkWindow_Add_100k(b *testing.B) {
	w := NewWindow(100_000, time.Second)
	for i := 0; i < b.N; i++ {
		w.Add(rand.Float64())
	}
}
