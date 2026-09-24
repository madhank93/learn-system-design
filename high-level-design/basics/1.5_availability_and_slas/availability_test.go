package main

import (
	"math"
	"testing"
	"time"
)

func TestFromNines(t *testing.T) {
	cases := []struct {
		nines    float64
		wantFrac float64
	}{
		{1, 0.9},
		{2, 0.99},
		{3, 0.999},
		{4, 0.9999},
		{5, 0.99999},
	}
	for _, c := range cases {
		a := FromNines(c.nines)
		if math.Abs(a.Fraction-c.wantFrac) > 1e-9 {
			t.Errorf("FromNines(%.0f).Fraction = %.6f, want %.6f", c.nines, a.Fraction, c.wantFrac)
		}
	}
}

func TestAnnualDowntime3Nines(t *testing.T) {
	a := FromNines(3) // 99.9%
	d := a.AnnualDowntime()
	// 99.9% → ~8.76 hours/year
	wantHours := 8.76
	gotHours := d.Hours()
	if math.Abs(gotHours-wantHours) > 0.1 {
		t.Errorf("3-nines annual downtime: got %.2f h, want %.2f h", gotHours, wantHours)
	}
}

func TestAnnualDowntime5Nines(t *testing.T) {
	a := FromNines(5) // 99.999%
	d := a.AnnualDowntime()
	// 99.999% → ~5.26 minutes/year
	if d > 6*time.Minute {
		t.Errorf("5-nines downtime too high: %v (want ~5.26 min)", d)
	}
	if d < 4*time.Minute {
		t.Errorf("5-nines downtime too low: %v (want ~5.26 min)", d)
	}
}

func TestSerialAvailabilityDegradesWithMore(t *testing.T) {
	a1 := SerialAvailability(0.999)
	a2 := SerialAvailability(0.999, 0.999)
	a3 := SerialAvailability(0.999, 0.999, 0.999)

	if a1 <= a2 || a2 <= a3 {
		t.Errorf("serial availability should degrade: %.4f > %.4f > %.4f", a1, a2, a3)
	}
}

func TestSerialThreeNines(t *testing.T) {
	// 3 × 99.9% in series ≈ 99.7%
	got := SerialAvailability(0.999, 0.999, 0.999)
	if math.Abs(got-0.997) > 0.001 {
		t.Errorf("3×99.9%% serial = %.4f, want ~0.997", got)
	}
}

func TestParallelAvailabilityImproves(t *testing.T) {
	for n := 1; n <= 4; n++ {
		a := ParallelAvailability(0.99, n)
		if n > 1 && a <= ParallelAvailability(0.99, n-1) {
			t.Errorf("parallel availability should improve with more replicas at n=%d", n)
		}
	}
}

func TestParallelAvailabilityApproachesOne(t *testing.T) {
	a := ParallelAvailability(0.99, 4)
	if a < 0.99999 {
		t.Errorf("4 replicas of 99%% should approach 99.999%%, got %.6f", a)
	}
}

func TestErrorBudget30Day(t *testing.T) {
	// 99.9% SLO over 30 days → ~43.2 minutes
	eb := ErrorBudget{SLOFraction: 0.999, Period: 30 * 24 * time.Hour}
	allowed := eb.Allowed()
	wantMin := 43.2
	gotMin := allowed.Minutes()
	if math.Abs(gotMin-wantMin) > 1 {
		t.Errorf("99.9%% 30-day error budget: got %.1f min, want %.1f min", gotMin, wantMin)
	}
}

func TestAvailabilityFromMTBF(t *testing.T) {
	// MTBF=720h (30 days), MTTR=1h → ~99.86%
	a := AvailabilityFromMTBF(720*time.Hour, 1*time.Hour)
	want := 720.0 / 721.0
	if math.Abs(a-want) > 0.0001 {
		t.Errorf("MTBF/MTTR availability: got %.4f, want %.4f", a, want)
	}
}

func TestCascadeImpact(t *testing.T) {
	// upstream=99.9%, dep=99.9%, 1 call → 99.8% effective
	got := CascadeImpact(0.999, 0.999, 1)
	want := 0.999 * 0.999
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("cascade 1 call: got %.5f want %.5f", got, want)
	}
}

func TestCascadeWorsensWith3Calls(t *testing.T) {
	one := CascadeImpact(0.999, 0.99, 1)
	three := CascadeImpact(0.999, 0.99, 3)
	if three >= one {
		t.Errorf("more dep calls should reduce availability: 1=%v 3=%v", one, three)
	}
}
