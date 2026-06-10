import {
  createEffect,
  createSignal,
  ErrorBoundary,
  onCleanup,
  Suspense,
} from "solid-js";
import DeclarativeMap from "~/components/Map";
import AddPanel from "~/components/AddPanel";
import { useGetGpss } from "~/queries/use-get-gpss";
import type { WsCoordinate } from "~/types";

const wsUrl = import.meta.env.VITE_WS_URL;

const Home = () => {
  let socket: WebSocket;
  const [liveUpdate, setLiveUpdate] = createSignal<WsCoordinate | null>(null);

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

  const gpss = useGetGpss();

  return (
    <ErrorBoundary fallback={(err) => <div>Error: {err.message}</div>}>
      <Suspense fallback={<div>Loading markers...</div>}>
        <div class="h-svh relative">
          <DeclarativeMap markers={gpss.data ?? []} liveUpdate={liveUpdate()} />
          <AddPanel />
        </div>
      </Suspense>
    </ErrorBoundary>
  );
};

export default Home;
