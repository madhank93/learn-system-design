import { Vehicle } from "./vehicle";
import { VehicleType } from "./vehicle.type";

export class Truck extends Vehicle {
  constructor(vehicleRegisterNumber: string) {
    super(vehicleRegisterNumber, VehicleType.Truck);
  }
}
