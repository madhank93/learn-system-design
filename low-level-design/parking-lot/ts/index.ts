import Admin from "./src/account/admin";
import CarSpot from "./src/parking-spot/car.spot";
import HandicappedSpot from "./src/parking-spot/handicapped.spot";
import ParkingSpot from "./src/parking-spot/parking.spot";
import { ParkingSpotType } from "./src/parking-spot/parking.spot.type";
import ParkingFloor from "./src/parking.floor";
import ParkingLot from "./src/parking.lot";
import Car from "./src/vehicles/car";
import { VehicleType } from "./src/vehicles/vehicle.type";

let car = new Car("TN0001234");

let carSpot: ParkingSpot = new CarSpot("B1");
carSpot.assignVehicleToSpot(car);

let parkingLot = ParkingLot.getInstance().getParkingSpot(VehicleType.Car);

const admin = new Admin("name", "pass");

const handicapSpot: Map<ParkingSpotType, Array<ParkingSpot>> = new Map<
  ParkingSpotType.Compact,
  Array<HandicappedSpot>
>();

const pf = new ParkingFloor("Floor1");

console.log(admin.addParkingFloor(pf));

console.log(admin.addParkingFloor(pf));
