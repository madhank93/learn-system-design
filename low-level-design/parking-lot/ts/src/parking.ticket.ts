import crypto from "node:crypto";
import { VehicleType } from "./vehicles/vehicle.type";

export class ParkingTicket {
  private parkingTicketID: string;
  private vehicleType: VehicleType;
  private vehicleRegisterNumber: string;
  private parkingSpotID: string;
  private parkingFloorID: string;
  private startTime!: Date;
  private endTime!: Date;
  private amount!: number;

  constructor(
    vehicleType: VehicleType,
    vehicleRegisterNumber: string,
    parkingFloorID: string,
    parkingSpotID: string
  ) {
    this.vehicleType = vehicleType;
    this.vehicleRegisterNumber = vehicleRegisterNumber;
    this.parkingFloorID = parkingFloorID;
    this.parkingSpotID = parkingSpotID;
    this.parkingTicketID = crypto.randomUUID();
  }

  public setStartTime(currentDateTime: Date = new Date()) {
    this.startTime = currentDateTime;
    return this;
  }

  public getStartTime() {
    return this.startTime;
  }

  public getEndTime() {
    return this.endTime;
  }

  public getAmount() {
    return this.amount;
  }

  public setEndTime(currentDateTime: Date = new Date()) {
    this.endTime = currentDateTime;
    return this;
  }

  public setAmount(amount: number) {
    this.amount = amount;
    return this;
  }

  public getVehicleType() {
    return this.vehicleType;
  }

  public getParkingSpotID() {
    return this.parkingSpotID;
  }

  public getParkingFloorID() {
    return this.parkingFloorID;
  }

  public getVehicleRegisterNumber() {
    return this.vehicleRegisterNumber;
  }
}
