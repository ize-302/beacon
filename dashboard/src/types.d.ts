type WsCoordinate = {
  gps_id: number;
  latitude: number;
  longitude: number;
  bearing: number;
  timestamp: number;
};

type Coordinate = {
  latitude: number;
  longitude: number;
  updated_at: string;
};

export type Gps = {
  id: number;
  sn: string;
  vehicle: { id: number; plate_number: string; created_at: string };
  last_coordinate: Coordinate | null;
  created_at: string;
};
