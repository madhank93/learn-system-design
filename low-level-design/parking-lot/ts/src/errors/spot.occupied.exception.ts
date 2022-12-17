import { VehicleType } from "../vehicles/vehicle.type";

export default class SpotOccupiedException extends Error {
  constructor(vehicleType: VehicleType) {
    super(`Spot is unavailable for the vehicle type: ${vehicleType}`);
  }
}
