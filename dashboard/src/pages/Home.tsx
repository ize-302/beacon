import { useEffect, useMemo, useState } from "react";
import { ErrorBoundary } from "react-error-boundary";
import DeclarativeMap from "~/components/Map";
import AddPanel from "~/components/AddPanel";
import { useGetVehicles } from "~/queries/use-get-vehicles";
import { useGetVehicleHistory } from "~/queries/use-get-vehicle-history";
import type { WsCoordinate, WsFrame } from "~/types";

const wsUrl = import.meta.env.VITE_WS_URL;

const Home = () => {
  const [liveUpdates, setLiveUpdates] = useState<WsCoordinate[] | null>(null);
  const [selectedVehicleId, setSelectedVehicleId] = useState<number | null>(
    null,
  );
  const [liveTail, setLiveTail] = useState<[number, number][]>([]);

  useEffect(() => {
    const socket = new WebSocket(wsUrl);
    socket.onmessage = (event) => {
      try {
        const frame: WsFrame = JSON.parse(event.data);
        if (frame.type !== "positions" || !frame.points?.length) return;
        setLiveUpdates(frame.points);
      } catch {
        console.error("WS parse error", event.data);
      }
    };
    socket.onerror = (error) => console.error("WebSocket Error:", error);
    return () => socket.close();
  }, []);

  // Reset live tail whenever the selected vehicle changes
  useEffect(() => {
    setLiveTail([]);
  }, [selectedVehicleId]);

  // Append incoming WS points to the tail when they belong to the selected
  // vehicle. A frame can carry several, so take every match in order.
  useEffect(() => {
    if (!liveUpdates?.length || selectedVehicleId === null) return;
    const mine = liveUpdates
      .filter((p) => p.vehicle_id === selectedVehicleId)
      .map((p) => [p.longitude, p.latitude] as [number, number]);
    if (mine.length) setLiveTail((prev) => [...prev, ...mine]);
  }, [liveUpdates, selectedVehicleId]);

  const vehicles = useGetVehicles();
  const history = useGetVehicleHistory(selectedVehicleId);

  // Initial history (oldest-first) + live tail appended as vehicle moves.
  // Memoized so a WS frame for a different vehicle (which changes
  // liveUpdates/Home's render but not history.data or liveTail) doesn't
  // force the map to redraw the route on every tick.
  const historyCoordinates = useMemo(() => {
    const fetched = history.data?.coordinates;
    const base = fetched?.length
      ? [...fetched]
          .reverse()
          .map((c) => [c.longitude, c.latitude] as [number, number])
      : [];
    const combined = [...base, ...liveTail];
    return combined.length ? combined : null;
  }, [history.data, liveTail]);

  if (vehicles.isLoading) {
    return <div>Loading markers...</div>;
  }

  return (
    <ErrorBoundary
      fallbackRender={({ error }) => (
        <div>Error: {error instanceof Error ? error.message : String(error)}</div>
      )}
    >
      <div className="h-svh relative">
        <DeclarativeMap
          markers={vehicles.data ?? []}
          liveUpdates={liveUpdates}
          onSelectVehicle={(id) =>
            setSelectedVehicleId((prev) => (prev === id ? null : id))
          }
          historyCoordinates={historyCoordinates}
        />

        <AddPanel vehicleCount={vehicles.data?.length ?? 0} />
      </div>
    </ErrorBoundary>
  );
};

export default Home;
