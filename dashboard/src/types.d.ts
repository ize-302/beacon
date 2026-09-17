export type WsCoordinate = {
  vehicle_id: number;
  latitude: number;
  longitude: number;
  bearing: number;
  timestamp: number;
};

// The socket sends one frame per write, carrying however many positions were
// recorded together. `type` discriminates it from the trip and driver events
// that will share this socket later.
export type WsFrame = {
  type: string;
  points: WsCoordinate[];
};
