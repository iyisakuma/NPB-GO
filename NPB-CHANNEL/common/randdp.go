package common

import "math"

// USE_POW ativa o uso de math.Pow (como se fosse #define USE_POW em C)
const USE_POW = true

var (
	r23, r46, t23, t46 float64
)

func init() {
	if USE_POW {
		r23 = math.Pow(0.5, 23.0)
		r46 = r23 * r23
		t23 = math.Pow(2.0, 23.0)
		t46 = t23 * t23
	} else {
		r23 = 1.0
		for i := 0; i < 23; i++ {
			r23 *= 0.5
		}
		r46 = r23 * r23

		t23 = 1.0
		for i := 0; i < 23; i++ {
			t23 *= 2.0
		}
		t46 = t23 * t23
	}
}
func Randlc(x *float64, a float64) float64 {
	var t1, t2, t3, t4, a1, a2, x1, x2, z float64

	t1 = r23 * a
	a1 = float64(int(t1))
	a2 = a - t23*a1

	t1 = r23 * (*x)
	x1 = float64(int(t1))
	x2 = *x - t23*x1

	t1 = a1*x2 + a2*x1
	t2 = float64(int(r23 * t1))
	z = t1 - t23*t2
	t3 = t23*z + a2*x2
	t4 = float64(int(r46 * t3))
	*x = t3 - t46*t4

	return r46 * (*x)
}

func Vranlc(n int, xSeed *float64, a float64, y []float64) {
	var t1, t2, t3, t4, a1, a2, x1, x2, z float64
	x := *xSeed

	t1 = r23 * a
	a1 = float64(int(t1))
	a2 = a - t23*a1

	for i := 0; i < n; i++ {
		t1 = r23 * x
		x1 = float64(int(t1))
		x2 = x - t23*x1

		t1 = a1*x2 + a2*x1
		t2 = float64(int(r23 * t1))
		z = t1 - t23*t2
		t3 = t23*z + a2*x2
		t4 = float64(int(r46 * t3))
		x = t3 - t46*t4
		y[i] = r46 * x
	}

	*xSeed = x
}
