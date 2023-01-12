import Admin from "./src/account/admin";
import { EntryPanel } from "./src/entry.panel";
import { ExitPanel } from "./src/exit.panel";
import { CarSpot } from "./src/parking-spot/car.spot";
import { ParkingFloor } from "./src/parking.floor";
import { Car } from "./src/vehicles/car";

const car = new Car("1234");
const carSpot = new CarSpot("F1-C1");
const floor1 = new ParkingFloor("Floor 1");
const admin = new Admin("name", "pass");

admin.addParkingFloor(floor1);
admin.addParkingSpot("Floor 1", carSpot);

const northEntrance = new EntryPanel("North");
const exitPanel = new ExitPanel();

floor1.showDisplayBoard();
const carParkingTicket = northEntrance.getParkingTicket(car);
floor1.showDisplayBoard();

exitPanel.checkout(carParkingTicket);
