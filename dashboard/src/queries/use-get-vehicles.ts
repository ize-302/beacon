import { useQuery } from "@tanstack/react-query";
import { vehiclesApi } from "~/api/client";

export const useGetVehicles = () => {
  return useQuery({
    queryKey: ["vehicles"],
    queryFn: async () => {
      const res = await vehiclesApi.getVehicles();
      return res.data.data;
    },
  });
};
