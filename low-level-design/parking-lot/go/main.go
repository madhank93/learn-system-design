package main

import (
	"fmt"
	"log"

	"parkinglot/parking"
)

func main() {
	lot := parking.GetInstance()
	admin := parking.NewAdmin("Jack", "Pass12345")

	// ── Setup ──────────────────────────────────────────────────
	floor1 := parking.NewParkingFloor("PF1001")
	floor2 := parking.NewParkingFloor("PF1002")
	must(admin.AddFloor(floor1))
	must(admin.AddFloor(floor2))

	ep1 := parking.NewEntryPanel("ENT001")
	ep2 := parking.NewEntryPanel("ENT002")
	xp1 := parking.NewExitPanel("EXT001")
	xp2 := parking.NewExitPanel("EXT002")
	must(admin.AddEntryPanel(ep1))
	must(admin.AddEntryPanel(ep2))
	must(admin.AddExitPanel(xp1))
	must(admin.AddExitPanel(xp2))

	// Add spots to floor 1
	must(admin.AddSpot("PF1001", parking.NewCarSpot("CPS1001")))
	must(admin.AddSpot("PF1001", parking.NewCarSpot("CPS1002")))
	must(admin.AddSpot("PF1001", parking.NewMotorcycleSpot("MSP1001")))
	must(admin.AddSpot("PF1001", parking.NewMotorcycleSpot("MSP1002")))
	must(admin.AddSpot("PF1001", parking.NewDisabledSpot("DSP1001")))

	// Add spots to floor 2
	must(admin.AddSpot("PF1002", parking.NewLargeSpot("LSP2001")))
	must(admin.AddSpot("PF1002", parking.NewElectricCarSpot("ECS2001")))

	fmt.Println("=== Display Board ===")
	floor1.ShowDisplayBoard()
	fmt.Println()

	// ── Park vehicles ──────────────────────────────────────────
	car1 := parking.NewCar("TN23890")
	car2 := parking.NewCar("TN99234")
	bike := parking.NewMotorcycle("TN34909")
	truck := parking.NewTruck("TN55001")

	entryPanel := lot.EntryPanels()[0]

	ticket1, err := entryPanel.GetTicket(car1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Car1 ticket: %s\n", ticket1)

	ticket2, err := entryPanel.GetTicket(car2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Car2 ticket: %s\n", ticket2)

	ticketBike, err := ep2.GetTicket(bike)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bike ticket: %s\n", ticketBike)

	ticketTruck, err := ep2.GetTicket(truck)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Truck ticket: %s\n", ticketTruck)

	fmt.Println("\n=== Display Board (after parking) ===")
	floor1.ShowDisplayBoard()
	floor2.ShowDisplayBoard()
	fmt.Println()

	// ── Checkout ───────────────────────────────────────────────
	exitPanel := lot.ExitPanels()[0]

	if err := exitPanel.Checkout(ticket1); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Car1 checked out. Amount: ₹%.0f\n", ticket1.Amount())

	if err := xp2.Checkout(ticketTruck); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Truck checked out. Amount: ₹%.0f\n", ticketTruck.Amount())

	fmt.Println("\n=== Display Board (after checkouts) ===")
	floor1.ShowDisplayBoard()
	floor2.ShowDisplayBoard()
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
