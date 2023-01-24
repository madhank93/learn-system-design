import { Vehicle } from "./vehicle";
import { VehicleType } from "./vehicle.type";

export class ElectricCar extends Vehicle {
  constructor(vehicleRegisterNumber: string) {
    super(vehicleRegisterNumber, VehicleType.ElectricCar);
  }
}
