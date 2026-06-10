import { createEffect, createSignal, onCleanup, onMount } from "solid-js";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import type { Gps, WsCoordinate } from "~/types";

mapboxgl.accessToken = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;

export default function DeclarativeMap(props: {
  markers: Gps[];
  liveUpdate: WsCoordinate | null;
}) {
  let mapContainer!: HTMLDivElement;
  let map: mapboxgl.Map;
  const markerInstances = new Map<number, mapboxgl.Marker>();
  const [mapReady, setMapReady] = createSignal(false);

  onMount(() => {
    map = new mapboxgl.Map({
      container: mapContainer,
      style: "mapbox://styles/mapbox/streets-v12",
      center: [3.37936, 6.5103],
      zoom: 8,
    });
    map.addControl(new mapboxgl.NavigationControl(), "top-right");
    map.on("load", () => setMapReady(true));
  });

  // Plot initial markers from REST response
  createEffect(() => {
    if (!mapReady() || !props.markers?.length) return;

    markerInstances.forEach((m) => m.remove());
    markerInstances.clear();

    props.markers.forEach((gps) => {
      const { longitude, latitude } = gps.last_coordinate;
      const popup = new mapboxgl.Popup({ offset: 25 }).setHTML(
        `<strong>${gps.sn}</strong><br/>${gps.vehicle.plate_number}`,
      );
      const marker = new mapboxgl.Marker()
        .setLngLat([longitude, latitude])
        .setPopup(popup)
        .addTo(map);
      markerInstances.set(gps.id, marker);
    });
  });

  // Move marker on WS update
  createEffect(() => {
    const update = props.liveUpdate;
    if (!update) return;
    const marker = markerInstances.get(update.gps_id);
    if (marker) marker.setLngLat([update.longitude, update.latitude]);
  });

  onCleanup(() => {
    markerInstances.forEach((m) => m.remove());
    if (map) map.remove();
  });

  return <div ref={mapContainer} style={{ width: "100%", height: "100%" }} />;
}
