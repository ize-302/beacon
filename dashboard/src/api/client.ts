import { Configuration, GpsApi, VehiclesApi, GpsPointsApi } from "~/client";

const config = new Configuration({
  basePath: import.meta.env.VITE_API_BASE_URL,
});

export const gpsApi = new GpsApi(config);
export const vehiclesApi = new VehiclesApi(config);
export const gpsPointsApi = new GpsPointsApi(config);
