package main

import (
	"fmt"
	"math"
	"time"
)

const hoursPerYear = 8760.0

// ── Availability nines ────────────────────────────────────────

// Availability holds uptime fraction and derived downtime.
type Availability struct {
	Nines    float64 // e.g. 3 = 99.9%
	Fraction float64 // 0.999
}

func FromNines(nines float64) Availability {
	f := 1 - math.Pow(10, -nines)
	return Availability{Nines: nines, Fraction: f}
}

func (a Availability) AnnualDowntime() time.Duration {
	downFraction := 1 - a.Fraction
	hours := downFraction * hoursPerYear
	return time.Duration(hours * float64(time.Hour))
}

func (a Availability) String() string {
	return fmt.Sprintf("%.3f%%", a.Fraction*100)
}

// ── System availability: serial vs parallel ───────────────────

// SerialAvailability: all components must be up.
// A_system = A1 × A2 × ... × An
func SerialAvailability(components ...float64) float64 {
	result := 1.0
	for _, a := range components {
		result *= a
	}
	return result
}

// ParallelAvailability: system up if at least one replica is up.
// A_system = 1 - (1-A)^n
func ParallelAvailability(singleAvailability float64, replicas int) float64 {
	return 1 - math.Pow(1-singleAvailability, float64(replicas))
}

// ── SLI / SLO / SLA ──────────────────────────────────────────

// ErrorBudget returns how much downtime is allowed per period
// given a target SLO availability fraction.
type ErrorBudget struct {
	SLOFraction float64
	Period      time.Duration
}

func (e ErrorBudget) Allowed() time.Duration {
	return time.Duration((1 - e.SLOFraction) * float64(e.Period))
}

func (e ErrorBudget) AllowedPercentage() float64 {
	return (1 - e.SLOFraction) * 100
}

// ── MTTR / MTBF ───────────────────────────────────────────────

// AvailabilityFromMTBF: A = MTBF / (MTBF + MTTR)
func AvailabilityFromMTBF(mtbf, mttr time.Duration) float64 {
	return mtbf.Seconds() / (mtbf.Seconds() + mttr.Seconds())
}

// ── Cascading failure amplification ──────────────────────────

// CascadeImpact models how a dependency's downtime amplifies upstream.
// If dep availability = A_dep, upstream call rate = r calls/req,
// upstream availability limited by: A_upstream * A_dep^r
func CascadeImpact(upstreamAvail, depAvail float64, callsPerRequest int) float64 {
	return upstreamAvail * math.Pow(depAvail, float64(callsPerRequest))
}

func main() {
	fmt.Println("=== The Nines Table ===")
	fmt.Printf("%-8s %-10s %-20s %-20s %-20s\n",
		"Nines", "Uptime", "Annual downtime", "Monthly downtime", "Weekly downtime")
	fmt.Println("────────────────────────────────────────────────────────────────────────────")
	for _, n := range []float64{1, 2, 3, 4, 5} {
		a := FromNines(n)
		annual := a.AnnualDowntime()
		monthly := time.Duration(float64(annual) / 12)
		weekly := time.Duration(float64(annual) / 52)
		fmt.Printf("%-8.0f %-10s %-20s %-20s %-20s\n",
			n, a, annual.Round(time.Second), monthly.Round(time.Second), weekly.Round(time.Second))
	}

	fmt.Println("\n=== Serial availability: A1 × A2 × A3 ===")
	components := []float64{0.999, 0.999, 0.999}
	serial := SerialAvailability(components...)
	fmt.Printf("  3 × 99.9%% services in series → %.4f%% (%.0f h/yr downtime)\n",
		serial*100, (1-serial)*hoursPerYear)

	fmt.Println("\n=== Parallel availability (adding replicas) ===")
	base := 0.99 // single 99% instance
	fmt.Printf("  Base: single instance at %.2f%%\n", base*100)
	for _, n := range []int{1, 2, 3, 4} {
		pa := ParallelAvailability(base, n)
		fmt.Printf("  %d replicas → %.5f%% (%.0f min/yr downtime)\n",
			n, pa*100, (1-pa)*hoursPerYear*60)
	}

	fmt.Println("\n=== Error Budget (30-day window) ===")
	for _, slo := range []float64{0.99, 0.999, 0.9999} {
		eb := ErrorBudget{SLOFraction: slo, Period: 30 * 24 * time.Hour}
		fmt.Printf("  SLO=%.4f%% → error budget = %s / month\n",
			slo*100, eb.Allowed().Round(time.Second))
	}

	fmt.Println("\n=== MTBF/MTTR → Availability ===")
	cases := []struct {
		mtbf, mttr time.Duration
	}{
		{720 * time.Hour, 1 * time.Hour},
		{720 * time.Hour, 4 * time.Hour},
		{168 * time.Hour, 30 * time.Minute},
	}
	for _, c := range cases {
		a := AvailabilityFromMTBF(c.mtbf, c.mttr)
		fmt.Printf("  MTBF=%s MTTR=%s → %.4f%%\n", c.mtbf, c.mttr, a*100)
	}

	fmt.Println("\n=== Cascading failure: effect of dependency availability ===")
	for _, depA := range []float64{0.999, 0.99, 0.95} {
		effective := CascadeImpact(0.999, depA, 3)
		fmt.Printf("  upstream=99.9%% dep=%.1f%% 3 calls/req → effective=%.3f%%\n",
			depA*100, effective*100)
	}
}
