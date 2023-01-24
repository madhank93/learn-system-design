import { VehicleType } from "./vehicle.type";

export abstract class Vehicle {
  private registerNumber: string;
  private type: VehicleType;

  constructor(vehicleRegisterNumber: string, vehicleType: VehicleType) {
    this.registerNumber = vehicleRegisterNumber;
    this.type = vehicleType;
  }

  public getVehicleType() {
    return this.type;
  }

  public getVehicleRegisterNumber() {
    return this.registerNumber;
  }
}
