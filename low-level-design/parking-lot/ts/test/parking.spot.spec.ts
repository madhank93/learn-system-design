import { beforeAll, beforeEach, describe, expect, it } from "vitest";
import { CarSpot } from "../src/parking-spot/car.spot";
import { MotorcycleSpot } from "../src/parking-spot/motorcycle.spot";
import { ParkingSpot } from "../src/parking-spot/parking.spot";
import { ParkingSpotType } from "../src/parking-spot/parking.spot.type";
import { Car } from "../src/vehicles/car";
import { MotorCycle } from "../src/vehicles/motor.cycle";
import { Vehicle } from "../src/vehicles/vehicle";
import { VehicleType } from "../src/vehicles/vehicle.type";

describe("Parking spots test cases", () => {
  let carSpot: ParkingSpot;
  let bikeSpot: ParkingSpot;

  beforeEach(() => {
    carSpot = new CarSpot("2F001");
    bikeSpot = new MotorcycleSpot("1F010");
  });

  it("Verify parking spot instances", () => {
    expect(carSpot).toBeInstanceOf(CarSpot);
    expect(bikeSpot).toBeInstanceOf(MotorcycleSpot);
  });

  it("Verify parking spot types", () => {
    expect(carSpot.getParkingSpotType()).toEqual(ParkingSpotType.Compact);
    expect(bikeSpot.getParkingSpotType()).toEqual(ParkingSpotType.Motorcycle);
  });

  it("Verify parking spot id", () => {
    expect(carSpot.getParkingSpotID()).toEqual("2F001");
    expect(bikeSpot.getParkingSpotID()).toEqual("1F010");
  });

  it("Verify parking spot availability when its free", () => {
    expect(carSpot.isSpotFree()).toEqual(true);
    expect(bikeSpot.isSpotFree()).toEqual(true);
  });

  it("Verify parking spot availability when its occupied", () => {
    const car = new Car("TN12345");
    const bike = new MotorCycle("TN10005");

    carSpot.assignVehicleToSpot(car);
    bikeSpot.assignVehicleToSpot(bike);

    expect(carSpot.isSpotFree()).toEqual(false);
    expect(bikeSpot.isSpotFree()).toEqual(false);
  });

  it("Verify parking spot availability when occupied spot is vacated", () => {
    const car = new Car("TN12345");
    const bike = new MotorCycle("TN10005");

    carSpot.assignVehicleToSpot(car);
    bikeSpot.assignVehicleToSpot(bike);

    carSpot.vacateVehicleFromSpot();
    bikeSpot.vacateVehicleFromSpot();

    expect(carSpot.isSpotFree()).toEqual(true);
    expect(bikeSpot.isSpotFree()).toEqual(true);
  });

  it("Verify able to park vehicle on the occupied spot", () => {
    const car = new Car("TN12345");
    carSpot.assignVehicleToSpot(car);

    expect(() => carSpot.assignVehicleToSpot(car)).toThrowError(
      "No spots are available for Car"
    );
  });

  it("Verify details of the vehicle parked in the spot", () => {
    const car = new Car("TN12345");
    carSpot.assignVehicleToSpot(car);
    const parkedVehicle = carSpot.getVehicleDetails()!;

    expect(parkedVehicle).toBeInstanceOf(Vehicle);
    expect(parkedVehicle.getVehicleRegisterNumber()).toEqual("TN12345");
    expect(parkedVehicle.getVehicleType()).toEqual(VehicleType.Car);
  });
});
