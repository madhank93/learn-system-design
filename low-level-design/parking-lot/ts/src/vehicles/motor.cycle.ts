import { Vehicle } from "./vehicle";
import { VehicleType } from "./vehicle.type";

export class MotorCycle extends Vehicle {
  constructor(vehicleRegisterNumber: string) {
    super(vehicleRegisterNumber, VehicleType.Motorcycle);
  }
}
