import Admin from "./src/account/admin";
import EntryPanel from "./src/entry.panel";
import ExitPanel from "./src/exit.panel";
import CarSpot from "./src/parking-spot/car.spot";
import HandicappedSpot from "./src/parking-spot/handicapped.spot";
import ParkingSpot from "./src/parking-spot/parking.spot";
import { ParkingSpotType } from "./src/parking-spot/parking.spot.type";
import ParkingFloor from "./src/parking.floor";
import ParkingLot from "./src/parking.lot";
import Car from "./src/vehicles/car";
import { VehicleType } from "./src/vehicles/vehicle.type";

const car = new Car("TN0001234");

const carSpot = new CarSpot("F1-C1");

const floor1 = new ParkingFloor("Floor 1");

const admin = new Admin("name", "pass");

admin.addParkingFloor(floor1);
admin.addParkingSpot("Floor 1", carSpot);
//carSpot.assignVehicleToSpot(car);

const northEntrance = new EntryPanel("North");
const exitPanel = new ExitPanel();

const carParkingTicket = northEntrance.getParkingTicket(car);
exitPanel.checkout(carParkingTicket);
