package parking

// Vehicle is satisfied by every concrete vehicle type.
type Vehicle interface {
	Type() VehicleType
	Number() string
}

// baseVehicle holds the shared state for all vehicle types.
type baseVehicle struct {
	vType  VehicleType
	number string
}

func (v *baseVehicle) Type() VehicleType { return v.vType }
func (v *baseVehicle) Number() string    { return v.number }

// Concrete vehicle types — each just injects its VehicleType constant.

type Car struct{ baseVehicle }
type Motorcycle struct{ baseVehicle }
type Truck struct{ baseVehicle }
type ElectricCar struct{ baseVehicle }
type Van struct{ baseVehicle }

func NewCar(number string) *Car {
	return &Car{baseVehicle{VehicleTypeCar, number}}
}
func NewMotorcycle(number string) *Motorcycle {
	return &Motorcycle{baseVehicle{VehicleTypeMotorcycle, number}}
}
func NewTruck(number string) *Truck {
	return &Truck{baseVehicle{VehicleTypeTruck, number}}
}
func NewElectricCar(number string) *ElectricCar {
	return &ElectricCar{baseVehicle{VehicleTypeElectricCar, number}}
}
func NewVan(number string) *Van {
	return &Van{baseVehicle{VehicleTypeVan, number}}
}
