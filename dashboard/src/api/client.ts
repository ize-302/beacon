import { Configuration, VehiclesApi, GPSPointsApi } from "~/client";

const config = new Configuration({
  basePath: import.meta.env.VITE_API_BASE_URL,
});

export const vehiclesApi = new VehiclesApi(config);
export const gpsPointsApi = new GPSPointsApi(config);
