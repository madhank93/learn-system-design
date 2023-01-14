import { beforeAll, describe, expect, expectTypeOf, it } from "vitest";
import { ParkingSpotType } from "../src/parking-spot/parking.spot.type";
import { ParkingFloor } from "../src/parking.floor";

describe("Parking floor test cases", () => {
  let firstFloor: ParkingFloor;
  let secondFloor: ParkingFloor;

  beforeAll(() => {
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

  it("Verify able to add parking spots to a floor", () => {});
});
