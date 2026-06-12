import {
  createEffect,
  createSignal,
  ErrorBoundary,
  onCleanup,
  Suspense,
} from "solid-js";
import DeclarativeMap from "~/components/Map";
import AddPanel from "~/components/AddPanel";
import { useGetGpsDevices } from "~/queries/use-get-gps-devices";
import { useGetGpsHistory } from "~/queries/use-get-gps-history";
import type { WsCoordinate } from "~/types";

const wsUrl = import.meta.env.VITE_WS_URL;

const Home = () => {
  let socket: WebSocket;
  const [liveUpdate, setLiveUpdate] = createSignal<WsCoordinate | null>(null);
  const [selectedGpsId, setSelectedGpsId] = createSignal<number | null>(null);
  const [liveTail, setLiveTail] = createSignal<[number, number][]>([]);

  createEffect(() => {
    socket = new WebSocket(wsUrl);
    socket.onmessage = (event) => {
      try {
        setLiveUpdate(JSON.parse(event.data));
      } catch {
        console.error("WS parse error", event.data);
      }
    };
    socket.onerror = (error) => console.error("WebSocket Error:", error);
    onCleanup(() => socket.close());
  });

  // Reset live tail whenever the selected vehicle changes
  createEffect(() => {
    selectedGpsId();
    setLiveTail([]);
  });

  // Append incoming WS point to the tail when it belongs to the selected vehicle
  createEffect(() => {
    const update = liveUpdate();
    if (!update || update.gps_id !== selectedGpsId()) return;
    setLiveTail((prev) => [...prev, [update.longitude, update.latitude]]);
  });

  const gpsDevices = useGetGpsDevices();
  const history = useGetGpsHistory(selectedGpsId);

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
            markers={gpsDevices.data ?? []}
            liveUpdate={liveUpdate()}
            onSelectGps={(id) =>
              setSelectedGpsId((prev) => (prev === id ? null : id))
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
