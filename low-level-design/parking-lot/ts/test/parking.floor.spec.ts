import { beforeEach, describe, expect, expectTypeOf, it } from "vitest";
import { CarSpot } from "../src/parking-spot/car.spot";
import { ParkingSpotType } from "../src/parking-spot/parking.spot.type";
import { ParkingFloor } from "../src/parking.floor";
import { Car } from "../src/vehicles/car";

describe("Parking floor test cases", () => {
  let firstFloor: ParkingFloor;
  let secondFloor: ParkingFloor;

  beforeEach(() => {
    firstFloor = new ParkingFloor("F001");
    secondFloor = new ParkingFloor("F002");
  });

  it("Verify the parking floor ID", () => {
    expect(firstFloor.getParkingFloorID()).toEqual("F001");
    expect(secondFloor.getParkingFloorID()).toEqual("F002");
  });

  it("Verify available parking types on a floor", () => {
    const spotTypesOfFirstFloor = [
      ...firstFloor.getListOfParkingSpots().keys(),
    ];
    const spotTypesOfSecondFloor = [
      ...secondFloor.getListOfParkingSpots().keys(),
    ];

    const expectedSpotTypes = [
      ParkingSpotType.Compact,
      ParkingSpotType.Motorcycle,
      ParkingSpotType.Large,
      ParkingSpotType.ElectricCar,
    ];
    expect(spotTypesOfFirstFloor).toEqual(expectedSpotTypes);
    expect(spotTypesOfSecondFloor).toEqual(expectedSpotTypes);
    expectTypeOf(spotTypesOfFirstFloor).items.toEqualTypeOf<ParkingSpotType>();
    expectTypeOf(spotTypesOfSecondFloor).items.toEqualTypeOf<ParkingSpotType>();
  });

  it("Verify initial size of the parking spot on the floor", () => {
    const spotTypesSizeOfFirstFloor = [
      ...firstFloor.getListOfParkingSpots().values(),
    ].flatMap((spot) => spot.length);
    const spotTypesSizeOfSecondFloor = [
      ...secondFloor.getListOfParkingSpots().values(),
    ].flatMap((spot) => spot.length);

    const expectedSpotTypesSize = [0, 0, 0, 0];

    expect(spotTypesSizeOfFirstFloor).toEqual(expectedSpotTypesSize);
    expect(spotTypesSizeOfSecondFloor).toEqual(expectedSpotTypesSize);
  });

  it("Verify able to add parking spots to a floor", () => {
    const carSpot = new CarSpot("PS1001");
    firstFloor
      .getListOfParkingSpots()
      .get(carSpot.getParkingSpotType())
      ?.push(carSpot);

    expect(
      firstFloor.getListOfParkingSpots().get(ParkingSpotType.Compact)?.length
    ).toEqual(1);
  });

  it("Verify it returns undefined when spots are unavailable for a vehicle", () => {
    const car = new Car("TN10010");
    const availableSpot = firstFloor.getAvailableSpot(car);

    expect(availableSpot).toBeUndefined();
  });

  it("Verify it returns available parking spot for a vehicle", () => {
    const car = new Car("TN10010");
    const firstCarSpot = new CarSpot("PS1001");
    const secondCarSpot = new CarSpot("PS1002");

    const carParkingSpot = firstFloor
      .getListOfParkingSpots()
      .get(firstCarSpot.getParkingSpotType());

    carParkingSpot?.push(firstCarSpot);
    carParkingSpot?.push(secondCarSpot);

    let availableSpot = firstFloor.getAvailableSpot(car);

    expect(availableSpot).toBeDefined();
    expect(availableSpot).toEqual(firstCarSpot);
    expect(availableSpot?.getParkingSpotID()).toEqual("PS1001");

    firstCarSpot.assignVehicleToSpot(car);
    availableSpot = firstFloor.getAvailableSpot(car);

    expect(availableSpot?.getParkingSpotID()).toEqual("PS1002");
  });

  it("Verify able to park the vehicle when spot is available", () => {
    const car = new Car("TN10010");
    const carSpot = new CarSpot("PS1001");
    firstFloor
      .getListOfParkingSpots()
      .get(carSpot.getParkingSpotType())
      ?.push(carSpot);

    const isSpotAvailable = firstFloor.canPark(car);

    expect(isSpotAvailable).true;
  });
});
