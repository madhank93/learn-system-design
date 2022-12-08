import Vehicle from "./vehicle";
import { VehicleType } from "./vehicle.type";

export default class Car extends Vehicle {
  constructor(vehicleRegisterNumber: string) {
    super(vehicleRegisterNumber, VehicleType.Car);
  }
}
