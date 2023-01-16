import Admin from "./src/account/admin";
import { EntryPanel } from "./src/entry.panel";
import { ExitPanel } from "./src/exit.panel";
import { CarSpot } from "./src/parking-spot/car.spot";
import { MotorcycleSpot } from "./src/parking-spot/motorcycle.spot";
import { ParkingFloor } from "./src/parking.floor";
import { ParkingLot } from "./src/parking.lot";
import { Car } from "./src/vehicles/car";
import { MotorCycle } from "./src/vehicles/motor.cycle";

// Crating parking lot instance
const parkingLot = ParkingLot.getInstance();

// Creating Admin account
const admin = new Admin("Jack", "Pass12345");

// Adding Parking floors
admin.addParkingFloor(new ParkingFloor("PF1001"));
admin.addParkingFloor(new ParkingFloor("PF1002"));

// Adding Entry panels
const entryPanel1 = new EntryPanel("ENT001");
const entryPanel2 = new EntryPanel("ENT002");
admin.addEntrancePanel(entryPanel1);
admin.addEntrancePanel(entryPanel2);

// Adding Exit panels
const exitPanel1 = new ExitPanel("EXT001");
const exitPanel2 = new ExitPanel("EXT002");
admin.addExitPanel(exitPanel1);
admin.addExitPanel(exitPanel2);

const firstFloorID = parkingLot
  .getListOfParkingFloor()
  .at(0)
  ?.getParkingFloorID();

// Adding Parking spots to the floor
const carParkingSpot1 = new CarSpot("CPS1001");
admin.addParkingSpot(firstFloorID!, carParkingSpot1);
const carParkingSpot2 = new CarSpot("CPS1002");
admin.addParkingSpot(firstFloorID!, carParkingSpot2);
const bikeParkingSpot1 = new MotorcycleSpot("MSP1001");
admin.addParkingSpot(firstFloorID!, bikeParkingSpot1);
const bikeParkingSpot2 = new MotorcycleSpot("MSP1002");
admin.addParkingSpot(firstFloorID!, bikeParkingSpot2);

// Creating vehicles
const car1 = new Car("TN23890");
const car2 = new Car("TN99234");
const bike = new MotorCycle("TN34909");

// Park the vehicle
let ticketForCar1 = entryPanel1.getParkingTicket(car1);
console.log("*** Parking ticket for car1 *** \n", ticketForCar1);

// Checkout the vehicle
ticketForCar1 = exitPanel2.checkout(ticketForCar1);
console.log("*** Parking ticket for car1 *** \n", ticketForCar1);
