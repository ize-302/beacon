import {
  createEffect,
  createSignal,
  ErrorBoundary,
  onCleanup,
  Suspense,
} from "solid-js";
import DeclarativeMap from "~/components/Map";
import AddPanel from "~/components/AddPanel";
import { useGetVehicles } from "~/queries/use-get-vehicles";
import { useGetVehicleHistory } from "~/queries/use-get-vehicle-history";
import type { WsCoordinate, WsFrame } from "~/types";

const wsUrl = import.meta.env.VITE_WS_URL;

const Home = () => {
  let socket: WebSocket;
  const [liveUpdates, setLiveUpdates] = createSignal<WsCoordinate[] | null>(
    null,
  );
  const [selectedVehicleId, setSelectedVehicleId] = createSignal<number | null>(
    null,
  );
  const [liveTail, setLiveTail] = createSignal<[number, number][]>([]);

  createEffect(() => {
    socket = new WebSocket(wsUrl);
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
    onCleanup(() => socket.close());
  });

  // Reset live tail whenever the selected vehicle changes
  createEffect(() => {
    selectedVehicleId();
    setLiveTail([]);
  });

  // Append incoming WS points to the tail when they belong to the selected
  // vehicle. A frame can carry several, so take every match in order.
  createEffect(() => {
    const frame = liveUpdates();
    const id = selectedVehicleId();
    if (!frame?.length || id === null) return;
    const mine = frame
      .filter((p) => p.vehicle_id === id)
      .map((p) => [p.longitude, p.latitude] as [number, number]);
    if (mine.length) setLiveTail((prev) => [...prev, ...mine]);
  });

  const vehicles = useGetVehicles();
  const history = useGetVehicleHistory(selectedVehicleId);

  // Initial history (oldest-first) + live tail appended as vehicle moves
  const historyCoordinates = () => {
    const fetched = history.data?.coordinates;
    const base = fetched?.length
      ? [...fetched]
          .reverse()
          .map((c) => [c.longitude, c.latitude] as [number, number])
      : [];
    const tail = liveTail();
    const combined = [...base, ...tail];
    return combined.length ? combined : null;
  };

  return (
    <ErrorBoundary fallback={(err) => <div>Error: {err.message}</div>}>
      <Suspense fallback={<div>Loading markers...</div>}>
        <div class="h-svh relative">
          <DeclarativeMap
            markers={vehicles.data ?? []}
            liveUpdates={liveUpdates()}
            onSelectVehicle={(id) =>
              setSelectedVehicleId((prev) => (prev === id ? null : id))
            }
            historyCoordinates={historyCoordinates()}
          />

          <AddPanel />
        </div>
      </Suspense>
    </ErrorBoundary>
  );
};

export default Home;
