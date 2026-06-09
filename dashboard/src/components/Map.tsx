import { onCleanup, onMount } from "solid-js";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";

const mapboxAccessToken = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;
mapboxgl.accessToken = mapboxAccessToken;

export default function DeclarativeMap() {
  let mapContainer!: HTMLDivElement;
  let map: mapboxgl.Map;

  onMount(() => {
    map = new mapboxgl.Map({
      container: mapContainer,
      style: "mapbox://styles/mapbox/streets-v12",
      center: [6.5103, 3.37936],
      zoom: 9,
    });

    map.addControl(new mapboxgl.NavigationControl(), "top-right");
  });

  onCleanup(() => {
    if (map) map.remove(); // Prevents memory leaks
  });

  return <div ref={mapContainer} style={{ width: "100%", height: "100%" }} />;
}
