import MotorcycleSpot from "./src/parking-spot/motorcycle.spot";
import ParkingSpot from "./src/parking-spot/parking.spot";
import ParkingFloor from "./src/parking.floor";
import Car from "./src/vehicles/car";

let bmwBike = new Car("TN0001234");

let bikeSpot: ParkingSpot = new MotorcycleSpot("B1");
bikeSpot.assignVehicleToSpot(bmwBike);

// let firstFloor = new ParkingFloor(5, "F1");

// firstFloor.addParkingSpot("B1", bikeSpot);
