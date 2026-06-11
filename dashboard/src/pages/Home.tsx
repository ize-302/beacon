import {
  createEffect,
  createSignal,
  ErrorBoundary,
  onCleanup,
  Show,
  Suspense,
} from "solid-js";
import DeclarativeMap from "~/components/Map";
import AddPanel from "~/components/AddPanel";
import { useGetGpss } from "~/queries/use-get-gpss";
import { useGetGpsHistory } from "~/queries/use-get-gps-history";
import type { WsCoordinate } from "~/types";
import { Card, CardContent, CardHeader } from "~/components/ui/card";

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

  const gpss = useGetGpss();
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

  const totalPoints = () => {
    const fetched = history.data?.coordinates.length ?? 0;
    return fetched + liveTail().length;
  };

  const selectedGps = () => gpss.data?.find((g) => g.id === selectedGpsId());

  return (
    <ErrorBoundary fallback={(err) => <div>Error: {err.message}</div>}>
      <Suspense fallback={<div>Loading markers...</div>}>
        <div class="h-svh relative">
          <DeclarativeMap
            markers={gpss.data ?? []}
            liveUpdate={liveUpdate()}
            onSelectGps={(id) =>
              setSelectedGpsId((prev) => (prev === id ? null : id))
            }
            historyCoordinates={historyCoordinates()}
          />

          <Show when={selectedGpsId() !== null}>
            <Card class="absolute bottom-6 left-4 z-10 w-64">
              <CardHeader class="flex-row py-1 items-center justify-between mb-2">
                <span class="font-semibold text-sm text-gray-800">
                  {selectedGps()?.sn ?? `GPS #${selectedGpsId()}`}
                </span>
                <button
                  class="text-gray-400 hover:text-gray-700 text-lg leading-none flex items-center justify-center"
                  onClick={() => setSelectedGpsId(null)}
                >
                  ✕
                </button>
              </CardHeader>
              <CardContent>
                <Show when={selectedGps()?.vehicle}>
                  <p class="text-xs text-gray-500 mb-2">
                    {selectedGps()!.vehicle!.plate_number}
                  </p>
                </Show>
                <Show when={history.isFetching}>
                  <p class="text-xs text-blue-500">Loading route…</p>
                </Show>
                <Show when={!history.isFetching && history.data}>
                  <p class="text-xs text-gray-600">
                    {totalPoints()} points tracked
                  </p>
                </Show>
                <Show when={history.isError}>
                  <p class="text-xs text-red-500">Failed to load history</p>
                </Show>
              </CardContent>
            </Card>
          </Show>

          <AddPanel />
        </div>
      </Suspense>
    </ErrorBoundary>
  );
};

export default Home;
