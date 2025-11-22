package main

import (
	"fmt"
	"math"
	"time"

	"github.com/iyisakuma/NPB-GO/NPB-SER/MG/params"
	"github.com/iyisakuma/NPB-GO/NPB-SER/common"
)

// Constants
const (
	LM         = 5
	LT_DEFAULT = 5
	NDIM1      = 5
	NDIM2      = 5
	NDIM3      = 5
	NM         = 2 + (1 << LM)
	ONE        = 1
	MAXLEVEL   = LT_DEFAULT + 1
	M          = NM + 1
	MM         = 10
	A          = 1220703125.0 // pow(5.0, 13.0)
	X          = 314159265.0
	T_INIT     = 0
	T_BENCH    = 1
	T_LAST     = 10
)

// MGBenchmark represents the MG (Multigrid) benchmark
type MGBenchmark struct {
	nx, ny, nz []int // Grid sizes for each level
	nit        int
	lt, lb     int // Level top and bottom
	class      string

	// Arrays - stored as flat arrays with offsets
	u, v, r    []float64
	a, c       []float64
	ir         []int
	m1, m2, m3 []int

	// Auxiliary arrays (Work buffers to avoid allocation in loops)
	// Size M is sufficient for all levels as M is max dimension
	u1, u2, r1, r2 []float64
	z1, z2, z3     []float64
	x1, y1         []float64

	// Setup variables
	is1, is2, is3 int // Start indices
	ie1, ie2, ie3 int // End indices
	n1, n2, n3    int // Actual array dimensions

	// Verification
	verified bool
	rnm2     float64
	rnmu     float64
}

// NewMGBenchmark creates a new MG benchmark instance
func NewMGBenchmark() *MGBenchmark {
	return &MGBenchmark{
		lt: LT_DEFAULT,
		lb: 1,
		nx: make([]int, MAXLEVEL+1),
		ny: make([]int, MAXLEVEL+1),
		nz: make([]int, MAXLEVEL+1),
		m1: make([]int, MAXLEVEL+1),
		m2: make([]int, MAXLEVEL+1),
		m3: make([]int, MAXLEVEL+1),
		ir: make([]int, MAXLEVEL+1),
		// Pre-allocate work buffers
		u1: make([]float64, M),
		u2: make([]float64, M),
		r1: make([]float64, M),
		r2: make([]float64, M),
		z1: make([]float64, M),
		z2: make([]float64, M),
		z3: make([]float64, M),
		x1: make([]float64, M),
		y1: make([]float64, M),
	}
}

// calculateIdx calculates 3D array index in a flat slice
func (mg *MGBenchmark) calculateIdx(i1, i2, i3, n1, n2 int) int {
	return i3*n2*n1 + i2*n1 + i1
}

// power raises an integer (disguised as double) to an integer power
func (mg *MGBenchmark) power(a float64, n int) float64 {
	power := 1.0
	nj := n
	aj := a

	for nj != 0 {
		if (nj % 2) == 1 {
			common.Randlc(&power, aj)
		}
		common.Randlc(&aj, aj)
		nj = nj / 2
	}

	return power
}

// zero3 zeros the first n elements of a slice
// CORREÇÃO CRÍTICA: Aceita 'n' para não zerar grids mais grossos acidentalmente
func zero3(z []float64, n int) {
	for i := 0; i < n; i++ {
		z[i] = 0.0
	}
}

// bubble does a bubble sort. Receives pointers to fixed arrays.
func (mg *MGBenchmark) bubble(ten *[2][MM]float64, j1, j2, j3 *[2][MM]int, m, ind int) {
	if ind == 1 {
		for i := 0; i < m-1; i++ {
			if ten[ind][i] > ten[ind][i+1] {
				// Swap
				ten[ind][i], ten[ind][i+1] = ten[ind][i+1], ten[ind][i]
				j1[ind][i], j1[ind][i+1] = j1[ind][i+1], j1[ind][i]
				j2[ind][i], j2[ind][i+1] = j2[ind][i+1], j2[ind][i]
				j3[ind][i], j3[ind][i+1] = j3[ind][i+1], j3[ind][i]
			} else {
				return
			}
		}
	} else {
		for i := 0; i < m-1; i++ {
			if ten[ind][i] < ten[ind][i+1] {
				// Swap
				ten[ind][i], ten[ind][i+1] = ten[ind][i+1], ten[ind][i]
				j1[ind][i], j1[ind][i+1] = j1[ind][i+1], j1[ind][i]
				j2[ind][i], j2[ind][i+1] = j2[ind][i+1], j2[ind][i]
				j3[ind][i], j3[ind][i+1] = j3[ind][i+1], j3[ind][i]
			} else {
				return
			}
		}
	}
}

// setup calculates grid sizes and offsets for all levels
func (mg *MGBenchmark) setup() {
	ng := make([][]int, MAXLEVEL+1)
	for i := range ng {
		ng[i] = make([]int, 3)
	}

	ng[mg.lt][0] = mg.nx[mg.lt]
	ng[mg.lt][1] = mg.ny[mg.lt]
	ng[mg.lt][2] = mg.nz[mg.lt]

	for ax := 0; ax < 3; ax++ {
		for k := mg.lt - 1; k >= 1; k-- {
			ng[k][ax] = ng[k+1][ax] / 2
		}
	}

	for k := mg.lt; k >= 1; k-- {
		mg.nx[k] = ng[k][0]
		mg.ny[k] = ng[k][1]
		mg.nz[k] = ng[k][2]
	}

	mi := make([][]int, MAXLEVEL+1)
	for i := range mi {
		mi[i] = make([]int, 3)
	}

	for k := mg.lt; k >= 1; k-- {
		for ax := 0; ax < 3; ax++ {
			mi[k][ax] = 2 + ng[k][ax]
		}
		mg.m1[k] = mi[k][0]
		mg.m2[k] = mi[k][1]
		mg.m3[k] = mi[k][2]
	}

	k := mg.lt
	mg.is1 = 2 + ng[k][0] - ng[mg.lt][0]
	mg.ie1 = 1 + ng[k][0]
	mg.n1 = 3 + mg.ie1 - mg.is1
	mg.is2 = 2 + ng[k][1] - ng[mg.lt][1]
	mg.ie2 = 1 + ng[k][1]
	mg.n2 = 3 + mg.ie2 - mg.is2
	mg.is3 = 2 + ng[k][2] - ng[mg.lt][2]
	mg.ie3 = 1 + ng[k][2]
	mg.n3 = 3 + mg.ie3 - mg.is3

	mg.ir[mg.lt] = 0
	for j := mg.lt - 1; j >= 1; j-- {
		mg.ir[j] = mg.ir[j+1] + ONE*mg.m1[j+1]*mg.m2[j+1]*mg.m3[j+1]
	}
}

// zran3 initializes grid with random values and places +1/-1 at extrema
func (mg *MGBenchmark) zran3(z []float64, n1, n2, n3 int, nx, ny int, k int) {
	a1 := mg.power(A, nx)
	a2 := mg.power(A, nx*ny)

	size := n3 * n2 * n1
	zero3(z, size)

	i := mg.is1 - 2 + nx*(mg.is2-2+ny*(mg.is3-2))
	ai := mg.power(A, i)

	d1 := mg.ie1 - mg.is1 + 1
	e2 := mg.ie2 - mg.is2 + 2
	e3 := mg.ie3 - mg.is3 + 2

	x0 := X
	common.Randlc(&x0, ai)

	for i3 := 1; i3 < e3; i3++ {
		x1 := x0
		for i2 := 1; i2 < e2; i2++ {
			xx := x1
			// Slice directly into z to avoid allocation
			startIdx := mg.calculateIdx(1, i2, i3, n1, n2)
			common.Vranlc(d1, &xx, A, z[startIdx:startIdx+d1])
			common.Randlc(&x1, a1)
		}
		common.Randlc(&x0, a2)
	}

	var ten [2][MM]float64
	var j1, j2, j3 [2][MM]int

	for i := 0; i < MM; i++ {
		ten[1][i] = 0.0
		ten[0][i] = 1.0
	}

	for i3 := 1; i3 < n3-1; i3++ {
		for i2 := 1; i2 < n2-1; i2++ {
			for i1 := 1; i1 < n1-1; i1++ {
				idx := mg.calculateIdx(i1, i2, i3, n1, n2)
				val := z[idx]

				if val > ten[1][0] {
					ten[1][0] = val
					j1[1][0] = i1
					j2[1][0] = i2
					j3[1][0] = i3
					mg.bubble(&ten, &j1, &j2, &j3, MM, 1)
				}
				if val < ten[0][0] {
					ten[0][0] = val
					j1[0][0] = i1
					j2[0][0] = i2
					j3[0][0] = i3
					mg.bubble(&ten, &j1, &j2, &j3, MM, 0)
				}
			}
		}
	}

	zero3(z, size)

	for i := MM - 1; i >= 0; i-- {
		idx := mg.calculateIdx(j1[0][i], j2[0][i], j3[0][i], n1, n2)
		z[idx] = -1.0
	}

	for i := MM - 1; i >= 0; i-- {
		idx := mg.calculateIdx(j1[1][i], j2[1][i], j3[1][i], n1, n2)
		z[idx] = 1.0
	}

	mg.comm3(z, n1, n2, n3, k)
}

func (mg *MGBenchmark) comm3(u []float64, n1, n2, n3 int, kk int) {
	// axis = 1
	for i3 := 1; i3 < n3-1; i3++ {
		for i2 := 1; i2 < n2-1; i2++ {
			idx0 := mg.calculateIdx(0, i2, i3, n1, n2)
			idx1 := mg.calculateIdx(n1-2, i2, i3, n1, n2)
			idx2 := mg.calculateIdx(n1-1, i2, i3, n1, n2)
			idx3 := mg.calculateIdx(1, i2, i3, n1, n2)
			u[idx0] = u[idx1]
			u[idx2] = u[idx3]
		}
	}
	// axis = 2
	for i3 := 1; i3 < n3-1; i3++ {
		for i1 := 0; i1 < n1; i1++ {
			idx0 := mg.calculateIdx(i1, 0, i3, n1, n2)
			idx1 := mg.calculateIdx(i1, n2-2, i3, n1, n2)
			idx2 := mg.calculateIdx(i1, n2-1, i3, n1, n2)
			idx3 := mg.calculateIdx(i1, 1, i3, n1, n2)
			u[idx0] = u[idx1]
			u[idx2] = u[idx3]
		}
	}
	// axis = 3
	for i2 := 0; i2 < n2; i2++ {
		for i1 := 0; i1 < n1; i1++ {
			idx0 := mg.calculateIdx(i1, i2, 0, n1, n2)
			idx1 := mg.calculateIdx(i1, i2, n3-2, n1, n2)
			idx2 := mg.calculateIdx(i1, i2, n3-1, n1, n2)
			idx3 := mg.calculateIdx(i1, i2, 1, n1, n2)
			u[idx0] = u[idx1]
			u[idx2] = u[idx3]
		}
	}
}

func (mg *MGBenchmark) norm2u3(r []float64, n1, n2, n3 int, nx, ny, nz int) float64 {
	dn := 1.0 * float64(nx*ny*nz)
	sum := 0.0
	rnmu := 0.0

	for i3 := 1; i3 < n3-1; i3++ {
		for i2 := 1; i2 < n2-1; i2++ {
			for i1 := 1; i1 < n1-1; i1++ {
				idx := mg.calculateIdx(i1, i2, i3, n1, n2)
				val := r[idx]
				sum += val * val
				a := math.Abs(val)
				if a > rnmu {
					rnmu = a
				}
			}
		}
	}

	mg.rnmu = rnmu
	return math.Sqrt(sum / dn)
}

func (mg *MGBenchmark) resid(u, v, r []float64, n1, n2, n3 int, a []float64, k int) {
	// Reuse pre-allocated buffers
	u1 := mg.u1[:n1]
	u2 := mg.u2[:n1]

	for i3 := 1; i3 < n3-1; i3++ {
		for i2 := 1; i2 < n2-1; i2++ {
			for i1 := 0; i1 < n1; i1++ {
				u1[i1] = u[mg.calculateIdx(i1, i2-1, i3, n1, n2)] +
					u[mg.calculateIdx(i1, i2+1, i3, n1, n2)] +
					u[mg.calculateIdx(i1, i2, i3-1, n1, n2)] +
					u[mg.calculateIdx(i1, i2, i3+1, n1, n2)]

				u2[i1] = u[mg.calculateIdx(i1, i2-1, i3-1, n1, n2)] +
					u[mg.calculateIdx(i1, i2+1, i3-1, n1, n2)] +
					u[mg.calculateIdx(i1, i2-1, i3+1, n1, n2)] +
					u[mg.calculateIdx(i1, i2+1, i3+1, n1, n2)]
			}

			for i1 := 1; i1 < n1-1; i1++ {
				idx := mg.calculateIdx(i1, i2, i3, n1, n2)
				r[idx] = v[idx] - a[0]*u[idx] -
					a[2]*(u2[i1]+u1[i1-1]+u1[i1+1]) -
					a[3]*(u2[i1-1]+u2[i1+1])
			}
		}
	}
	mg.comm3(r, n1, n2, n3, k)
}

func (mg *MGBenchmark) psinv(r, u []float64, n1, n2, n3 int, c []float64, k int) {
	r1 := mg.r1[:n1]
	r2 := mg.r2[:n1]

	for i3 := 1; i3 < n3-1; i3++ {
		for i2 := 1; i2 < n2-1; i2++ {
			for i1 := 0; i1 < n1; i1++ {
				r1[i1] = r[mg.calculateIdx(i1, i2-1, i3, n1, n2)] +
					r[mg.calculateIdx(i1, i2+1, i3, n1, n2)] +
					r[mg.calculateIdx(i1, i2, i3-1, n1, n2)] +
					r[mg.calculateIdx(i1, i2, i3+1, n1, n2)]

				r2[i1] = r[mg.calculateIdx(i1, i2-1, i3-1, n1, n2)] +
					r[mg.calculateIdx(i1, i2+1, i3-1, n1, n2)] +
					r[mg.calculateIdx(i1, i2-1, i3+1, n1, n2)] +
					r[mg.calculateIdx(i1, i2+1, i3+1, n1, n2)]
			}

			for i1 := 1; i1 < n1-1; i1++ {
				idx := mg.calculateIdx(i1, i2, i3, n1, n2)
				u[idx] = u[idx] + c[0]*r[idx] +
					c[1]*(r[idx-1]+r[idx+1]+r1[i1]) +
					c[2]*(r2[i1]+r1[i1-1]+r1[i1+1])
			}
		}
	}
	mg.comm3(u, n1, n2, n3, k)
}

func (mg *MGBenchmark) rprj3(r []float64, m1k, m2k, m3k int, s []float64, m1j, m2j, m3j int, k int) {
	var d1, d2, d3 int
	if m1k == 3 {
		d1 = 2
	} else {
		d1 = 1
	}
	if m2k == 3 {
		d2 = 2
	} else {
		d2 = 1
	}
	if m3k == 3 {
		d3 = 2
	} else {
		d3 = 1
	}

	x1 := mg.x1[:m1k]
	y1 := mg.y1[:m1k]

	for j3 := 1; j3 < m3j-1; j3++ {
		i3 := 2*j3 - d3
		for j2 := 1; j2 < m2j-1; j2++ {
			i2 := 2*j2 - d2
			for j1 := 1; j1 < m1j; j1++ {
				i1 := 2*j1 - d1
				x1[i1] = r[mg.calculateIdx(i1, i2, i3+1, m1k, m2k)] +
					r[mg.calculateIdx(i1, i2+2, i3+1, m1k, m2k)] +
					r[mg.calculateIdx(i1, i2+1, i3, m1k, m2k)] +
					r[mg.calculateIdx(i1, i2+1, i3+2, m1k, m2k)]

				y1[i1] = r[mg.calculateIdx(i1, i2, i3, m1k, m2k)] +
					r[mg.calculateIdx(i1, i2, i3+2, m1k, m2k)] +
					r[mg.calculateIdx(i1, i2+2, i3, m1k, m2k)] +
					r[mg.calculateIdx(i1, i2+2, i3+2, m1k, m2k)]
			}

			for j1 := 1; j1 < m1j-1; j1++ {
				i1 := 2*j1 - d1
				y2 := r[mg.calculateIdx(i1+1, i2, i3, m1k, m2k)] +
					r[mg.calculateIdx(i1+1, i2, i3+2, m1k, m2k)] +
					r[mg.calculateIdx(i1+1, i2+2, i3, m1k, m2k)] +
					r[mg.calculateIdx(i1+1, i2+2, i3+2, m1k, m2k)]

				x2 := r[mg.calculateIdx(i1+1, i2, i3+1, m1k, m2k)] +
					r[mg.calculateIdx(i1+1, i2+2, i3+1, m1k, m2k)] +
					r[mg.calculateIdx(i1+1, i2+1, i3, m1k, m2k)] +
					r[mg.calculateIdx(i1+1, i2+1, i3+2, m1k, m2k)]

				ridx := mg.calculateIdx(i1+1, i2+1, i3+1, m1k, m2k)
				sidx := mg.calculateIdx(j1, j2, j3, m1j, m2j)

				s[sidx] = 0.5*r[ridx] + 0.25*(r[ridx-1]+r[ridx+1]+x2) + 0.125*(x1[i1]+x1[i1+2]+y2) + 0.0625*(y1[i1]+y1[i1+2])
			}
		}
	}
	mg.comm3(s, m1j, m2j, m3j, k-1)
}

func (mg *MGBenchmark) interp(z []float64, mm1, mm2, mm3 int, u []float64, n1, n2, n3 int, k int) {
	var d1, d2, d3, t1, t2, t3 int

	if n1 != 3 && n2 != 3 && n3 != 3 {
		z1 := mg.z1[:mm1]
		z2 := mg.z2[:mm1]
		z3 := mg.z3[:mm1]

		for i3 := 0; i3 < mm3-1; i3++ {
			for i2 := 0; i2 < mm2-1; i2++ {
				for i1 := 0; i1 < mm1; i1++ {
					idx1 := mg.calculateIdx(i1, i2+1, i3, mm1, mm2)
					idx2 := mg.calculateIdx(i1, i2, i3, mm1, mm2)
					z1[i1] = z[idx1] + z[idx2]

					idx3 := mg.calculateIdx(i1, i2, i3+1, mm1, mm2)
					z2[i1] = z[idx3] + z[idx2]

					idx4 := mg.calculateIdx(i1, i2+1, i3+1, mm1, mm2)
					z3[i1] = z[idx4] + z[idx1] + z1[i1]
				}

				for i1 := 0; i1 < mm1-1; i1++ {
					zidx := mg.calculateIdx(i1, i2, i3, mm1, mm2)
					u[mg.calculateIdx(2*i1, 2*i2, 2*i3, n1, n2)] += z[zidx]
					u[mg.calculateIdx(2*i1+1, 2*i2, 2*i3, n1, n2)] += 0.5 * (z[mg.calculateIdx(i1+1, i2, i3, mm1, mm2)] + z[zidx])
				}
				for i1 := 0; i1 < mm1-1; i1++ {
					u[mg.calculateIdx(2*i1, 2*i2+1, 2*i3, n1, n2)] += 0.5 * z1[i1]
					u[mg.calculateIdx(2*i1+1, 2*i2+1, 2*i3, n1, n2)] += 0.25 * (z1[i1] + z1[i1+1])
				}
				for i1 := 0; i1 < mm1-1; i1++ {
					u[mg.calculateIdx(2*i1, 2*i2, 2*i3+1, n1, n2)] += 0.5 * z2[i1]
					u[mg.calculateIdx(2*i1+1, 2*i2, 2*i3+1, n1, n2)] += 0.25 * (z2[i1] + z2[i1+1])
				}
				for i1 := 0; i1 < mm1-1; i1++ {
					u[mg.calculateIdx(2*i1, 2*i2+1, 2*i3+1, n1, n2)] += 0.25 * z3[i1]
					u[mg.calculateIdx(2*i1+1, 2*i2+1, 2*i3+1, n1, n2)] += 0.125 * (z3[i1] + z3[i1+1])
				}
			}
		}
	} else {
		if n1 == 3 {
			d1, t1 = 2, 1
		} else {
			d1, t1 = 1, 0
		}
		if n2 == 3 {
			d2, t2 = 2, 1
		} else {
			d2, t2 = 1, 0
		}
		if n3 == 3 {
			d3, t3 = 2, 1
		} else {
			d3, t3 = 1, 0
		}

		for i3 := d3; i3 <= mm3-1; i3++ {
			for i2 := d2; i2 <= mm2-1; i2++ {
				for i1 := d1; i1 <= mm1-1; i1++ {
					zidx := mg.calculateIdx(i1-1, i2-1, i3-1, mm1, mm2)
					uidx := mg.calculateIdx(2*i3-d3-1, 2*i2-d2-1, 2*i1-d1-1, n1, n2)
					u[uidx] += z[zidx]
				}
				for i1 := 1; i1 <= mm1-1; i1++ {
					zidx1 := mg.calculateIdx(i1, i2-1, i3-1, mm1, mm2)
					zidx2 := mg.calculateIdx(i1-1, i2-1, i3-1, mm1, mm2)
					uidx := mg.calculateIdx(2*i3-d3-1, 2*i2-d2-1, 2*i1-t1-1, n1, n2)
					u[uidx] += 0.5 * (z[zidx1] + z[zidx2])
				}
			}
			for i2 := 1; i2 <= mm2-1; i2++ {
				for i1 := d1; i1 <= mm1-1; i1++ {
					zidx1 := mg.calculateIdx(i1-1, i2, i3-1, mm1, mm2)
					zidx2 := mg.calculateIdx(i1-1, i2-1, i3-1, mm1, mm2)
					uidx := mg.calculateIdx(2*i3-d3-1, 2*i2-t2-1, 2*i1-d1-1, n1, n2)
					u[uidx] += 0.5 * (z[zidx1] + z[zidx2])
				}
				for i1 := 1; i1 <= mm1-1; i1++ {
					zidx1 := mg.calculateIdx(i1, i2, i3-1, mm1, mm2)
					zidx2 := mg.calculateIdx(i1-1, i2, i3-1, mm1, mm2)
					zidx3 := mg.calculateIdx(i1, i2-1, i3-1, mm1, mm2)
					zidx4 := mg.calculateIdx(i1-1, i2-1, i3-1, mm1, mm2)
					uidx := mg.calculateIdx(2*i3-d3-1, 2*i2-t2-1, 2*i1-t1-1, n1, n2)
					u[uidx] += 0.25 * (z[zidx1] + z[zidx2] + z[zidx3] + z[zidx4])
				}
			}
		}

		for i3 := 1; i3 <= mm3-1; i3++ {
			for i2 := d2; i2 <= mm2-1; i2++ {
				for i1 := d1; i1 <= mm1-1; i1++ {
					zidx1 := mg.calculateIdx(i1-1, i2-1, i3, mm1, mm2)
					zidx2 := mg.calculateIdx(i1-1, i2-1, i3-1, mm1, mm2)
					uidx := mg.calculateIdx(2*i3-t3-1, 2*i2-d2-1, 2*i1-d1-1, n1, n2)
					u[uidx] += 0.5 * (z[zidx1] + z[zidx2])
				}
				for i1 := 1; i1 <= mm1-1; i1++ {
					zidx1 := mg.calculateIdx(i1, i2-1, i3, mm1, mm2)
					zidx2 := mg.calculateIdx(i1-1, i2-1, i3, mm1, mm2)
					zidx3 := mg.calculateIdx(i1, i2-1, i3-1, mm1, mm2)
					zidx4 := mg.calculateIdx(i1-1, i2-1, i3-1, mm1, mm2)
					uidx := mg.calculateIdx(2*i3-t3-1, 2*i2-d2-1, 2*i1-t1-1, n1, n2)
					u[uidx] += 0.25 * (z[zidx1] + z[zidx2] + z[zidx3] + z[zidx4])
				}
			}
			for i2 := 1; i2 <= mm2-1; i2++ {
				for i1 := d1; i1 <= mm1-1; i1++ {
					zidx1 := mg.calculateIdx(i1-1, i2, i3, mm1, mm2)
					zidx2 := mg.calculateIdx(i1-1, i2-1, i3, mm1, mm2)
					zidx3 := mg.calculateIdx(i1-1, i2, i3-1, mm1, mm2)
					zidx4 := mg.calculateIdx(i1-1, i2-1, i3-1, mm1, mm2)
					uidx := mg.calculateIdx(2*i3-t3-1, 2*i2-t2-1, 2*i1-d1-1, n1, n2)
					u[uidx] += 0.25 * (z[zidx1] + z[zidx2] + z[zidx3] + z[zidx4])
				}
				for i1 := 1; i1 <= mm1-1; i1++ {
					u[mg.calculateIdx(2*i3-t3-1, 2*i2-t2-1, 2*i1-t1-1, n1, n2)] += 0.125 *
						(z[mg.calculateIdx(i1, i2, i3, mm1, mm2)] +
							z[mg.calculateIdx(i1, i2-1, i3, mm1, mm2)] +
							z[mg.calculateIdx(i1, i2, i3-1, mm1, mm2)] +
							z[mg.calculateIdx(i1, i2-1, i3-1, mm1, mm2)] +
							z[mg.calculateIdx(i1-1, i2, i3, mm1, mm2)] +
							z[mg.calculateIdx(i1-1, i2-1, i3, mm1, mm2)] +
							z[mg.calculateIdx(i1-1, i2, i3-1, mm1, mm2)] +
							z[mg.calculateIdx(i1-1, i2-1, i3-1, mm1, mm2)])
				}
			}
		}
	}
}

func (mg *MGBenchmark) mg3P(u, v, r []float64, a, c []float64, n1, n2, n3 int, k int) {
	for k := mg.lt; k >= mg.lb+1; k-- {
		j := k - 1
		rk := mg.r[mg.ir[k]:]
		rj := mg.r[mg.ir[j]:]
		mg.rprj3(rk, mg.m1[k], mg.m2[k], mg.m3[k], rj, mg.m1[j], mg.m2[j], mg.m3[j], k)
	}

	k = mg.lb
	uk := mg.u[mg.ir[k]:]
	rk := mg.r[mg.ir[k]:]

	// CRITICAL FIX: Only zero the size of this level
	sizeK := mg.m1[k] * mg.m2[k] * mg.m3[k]
	zero3(uk, sizeK)

	mg.psinv(rk, uk, mg.m1[k], mg.m2[k], mg.m3[k], c, k)

	for k = mg.lb + 1; k <= mg.lt-1; k++ {
		j := k - 1
		uk := mg.u[mg.ir[k]:]
		uj := mg.u[mg.ir[j]:]
		rk := mg.r[mg.ir[k]:]

		// CRITICAL FIX: Only zero the size of this level
		sizeK = mg.m1[k] * mg.m2[k] * mg.m3[k]
		zero3(uk, sizeK)

		mg.interp(uj, mg.m1[j], mg.m2[j], mg.m3[j], uk, mg.m1[k], mg.m2[k], mg.m3[k], k)
		mg.resid(uk, rk, rk, mg.m1[k], mg.m2[k], mg.m3[k], a, k)
		mg.psinv(rk, uk, mg.m1[k], mg.m2[k], mg.m3[k], c, k)
	}

	j := mg.lt - 1
	k = mg.lt
	uj := mg.u[mg.ir[j]:]
	mg.interp(uj, mg.m1[j], mg.m2[j], mg.m3[j], u, n1, n2, n3, k)
	mg.resid(u, v, r, n1, n2, n3, a, k)
	mg.psinv(r, u, n1, n2, n3, c, k)
}

func (mg *MGBenchmark) run() {
	NV := ONE * (2 + (1 << NDIM1)) * (2 + (1 << NDIM2)) * (2 + (1 << NDIM3))
	NR := ((NV + NM*NM + 5*NM + 7*LM + 6) / 7) * 8

	// Allocations
	if mg.u == nil {
		mg.u = make([]float64, NR)
	}
	if mg.v == nil {
		mg.v = make([]float64, NV)
	}
	if mg.r == nil {
		mg.r = make([]float64, NR)
	}
	if mg.a == nil {
		mg.a = make([]float64, 4)
	}
	if mg.c == nil {
		mg.c = make([]float64, 4)
	}

	mg.a[0] = -8.0 / 3.0
	mg.a[1] = 0.0
	mg.a[2] = 1.0 / 6.0
	mg.a[3] = 1.0 / 12.0

	if mg.class == "A" || mg.class == "S" || mg.class == "W" {
		mg.c[0] = -3.0 / 8.0
		mg.c[1] = 1.0 / 32.0
		mg.c[2] = -1.0 / 64.0
		mg.c[3] = 0.0
	} else {
		mg.c[0] = -3.0 / 17.0
		mg.c[1] = 1.0 / 33.0
		mg.c[2] = -1.0 / 61.0
		mg.c[3] = 0.0
	}

	mg.setup()

	// Initialize arrays. Using len(mg.u) here is safe as it's the first init.
	zero3(mg.u, len(mg.u))
	mg.zran3(mg.v, mg.n1, mg.n2, mg.n3, mg.nx[mg.lt], mg.ny[mg.lt], mg.lt)

	mg.rnm2 = mg.norm2u3(mg.v, mg.n1, mg.n2, mg.n3, mg.nx[mg.lt], mg.ny[mg.lt], mg.nz[mg.lt])

	fmt.Printf("\n\n NAS Parallel Benchmarks 4.1 Serial Go version - MG Benchmark\n\n")
	fmt.Printf(" Size: %3dx%3dx%3d (class %s)\n", mg.nx[mg.lt], mg.ny[mg.lt], mg.nz[mg.lt], mg.class)
	fmt.Printf(" Iterations: %3d\n", mg.nit)

	mg.resid(mg.u, mg.v, mg.r, mg.n1, mg.n2, mg.n3, mg.a, mg.lt)
	mg.rnm2 = mg.norm2u3(mg.r, mg.n1, mg.n2, mg.n3, mg.nx[mg.lt], mg.ny[mg.lt], mg.nz[mg.lt])

	// Warm-up
	mg.mg3P(mg.u, mg.v, mg.r, mg.a, mg.c, mg.n1, mg.n2, mg.n3, mg.lt)
	mg.resid(mg.u, mg.v, mg.r, mg.n1, mg.n2, mg.n3, mg.a, mg.lt)

	mg.setup()

	zero3(mg.u, len(mg.u))
	mg.zran3(mg.v, mg.n1, mg.n2, mg.n3, mg.nx[mg.lt], mg.ny[mg.lt], mg.lt)

	mg.resid(mg.u, mg.v, mg.r, mg.n1, mg.n2, mg.n3, mg.a, mg.lt)
	mg.rnm2 = mg.norm2u3(mg.r, mg.n1, mg.n2, mg.n3, mg.nx[mg.lt], mg.ny[mg.lt], mg.nz[mg.lt])

	startTime := time.Now()
	for it := 1; it <= mg.nit; it++ {
		mg.mg3P(mg.u, mg.v, mg.r, mg.a, mg.c, mg.n1, mg.n2, mg.n3, mg.lt)
		mg.resid(mg.u, mg.v, mg.r, mg.n1, mg.n2, mg.n3, mg.a, mg.lt)

	}
	elapsed := time.Since(startTime).Seconds()

	mg.rnm2 = mg.norm2u3(mg.r, mg.n1, mg.n2, mg.n3, mg.nx[mg.lt], mg.ny[mg.lt], mg.nz[mg.lt])

	epsilon := 1.0e-8
	verifyValue := params.VERIFY_VALUE
	err := math.Abs(mg.rnm2-verifyValue) / verifyValue
	mg.verified = err <= epsilon

	fmt.Printf("\n Benchmark completed\n")
	if mg.verified {
		fmt.Printf(" VERIFICATION SUCCESSFUL\n")
		fmt.Printf(" L2 Norm is %20.13e\n", mg.rnm2)
		fmt.Printf(" Error is   %20.13e\n", err)
	} else {
		fmt.Printf(" VERIFICATION FAILED\n")
		fmt.Printf(" L2 Norm is             %20.13e\n", mg.rnm2)
		fmt.Printf(" The correct L2 Norm is %20.13e\n", verifyValue)
	}

	mops := 0.0
	if elapsed > 0 {
		nn := float64(mg.nx[mg.lt] * mg.ny[mg.lt] * mg.nz[mg.lt])
		mops = 58.0 * float64(mg.nit) * nn * 1.0e-6 / elapsed
	}

	common.PrintResults("MG", mg.class, mg.nx[mg.lt], mg.ny[mg.lt], mg.nz[mg.lt], mg.nit, elapsed, mops, "floating point", mg.verified, "4.1", "Unknown", "Go", "")
}
