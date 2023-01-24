import { beforeAll, describe, expect, it } from "vitest";

import { Car } from "../src/vehicles/car";
import { MotorCycle } from "../src/vehicles/motor.cycle";
import { Vehicle } from "../src/vehicles/vehicle";
import { VehicleType } from "../src/vehicles/vehicle.type";

describe("Vehicle specify test cases", () => {
  let car: Vehicle;
  let bike: Vehicle;

  beforeAll(() => {
    car = new Car("TN12345");
    bike = new MotorCycle("TN10005");
  });

  it("Verify vehicle instance type", () => {
    expect(car).toBeInstanceOf(Car);
    expect(bike).toBeInstanceOf(MotorCycle);
  });

  it("Verify vehicle register number", () => {
    expect(car.getVehicleRegisterNumber()).toEqual("TN12345");
    expect(bike.getVehicleRegisterNumber()).toEqual("TN10005");
  });

  it("Verify vehicle type", () => {
    expect(car.getVehicleType()).toEqual(VehicleType.Car);
    expect(bike.getVehicleType()).toEqual(VehicleType.Motorcycle);
  });
});
