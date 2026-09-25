package parking

// VehicleType identifies the kind of vehicle.
type VehicleType string

const (
	VehicleTypeCar         VehicleType = "Car"
	VehicleTypeMotorcycle  VehicleType = "Motorcycle"
	VehicleTypeTruck       VehicleType = "Truck"
	VehicleTypeElectricCar VehicleType = "ElectricCar"
	VehicleTypeVan         VehicleType = "Van"
)

// SpotType identifies the kind of parking spot.
type SpotType string

const (
	SpotTypeCompact     SpotType = "Compact"
	SpotTypeLarge       SpotType = "Large"
	SpotTypeDisabled    SpotType = "Disabled"
	SpotTypeMotorcycle  SpotType = "Motorcycle"
	SpotTypeElectricCar SpotType = "ElectricCar"
)

// vehicleToSpotType maps a vehicle type to the correct spot type.
func vehicleToSpotType(vt VehicleType) SpotType {
	switch vt {
	case VehicleTypeCar:
		return SpotTypeCompact
	case VehicleTypeMotorcycle:
		return SpotTypeMotorcycle
	case VehicleTypeElectricCar:
		return SpotTypeElectricCar
	default: // Truck, Van
		return SpotTypeLarge
	}
}
