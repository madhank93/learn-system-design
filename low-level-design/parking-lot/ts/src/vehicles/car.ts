import { Vehicle } from "./vehicle";
import { VehicleType } from "./vehicle.type";

export class Car extends Vehicle {
  constructor(vehicleRegisterNumber: string) {
    super(vehicleRegisterNumber, VehicleType.Car);
  }
}
