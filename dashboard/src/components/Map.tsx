import { createEffect, createSignal, onCleanup, onMount } from "solid-js";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import type { GpsGpsResponse } from "~/client/api";
import type { WsCoordinate } from "~/types";
import policeCarUrl from "~/components/vehicles/police-car.svg?url";

const vehicleIcons = [policeCarUrl];
const DEFAULT_ANIM_DURATION = 4000;
const HISTORY_SOURCE = "gps-history";
const HISTORY_LAYER = "gps-history-line";

mapboxgl.accessToken = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;

function makeMarkerEl(width: string, height: string, iconIndex: number, onClick?: () => void) {
  const el = document.createElement("img");
  el.src = vehicleIcons[iconIndex];
  el.style.width = width;
  el.style.height = height;
  el.style.cursor = "pointer";
  if (onClick) el.addEventListener("click", onClick);
  return el;
}

export default function DeclarativeMap(props: {
  markers: GpsGpsResponse[];
  liveUpdate: WsCoordinate | null;
  onSelectGps: (id: number) => void;
  historyCoordinates: [number, number][] | null;
}) {
  let mapContainer!: HTMLDivElement;
  let map: mapboxgl.Map;
  const markerInstances = new Map<number, mapboxgl.Marker>();
  const markerAnimations = new Map<number, number>();
  const markerTimestamps = new Map<number, number>();
  const [mapReady, setMapReady] = createSignal(false);

  function animateMarker(marker: mapboxgl.Marker, id: number, toLng: number, toLat: number, bearing: number, timestamp: number) {
    const prev = markerAnimations.get(id);
    if (prev !== undefined) cancelAnimationFrame(prev);

    const lastTs = markerTimestamps.get(id);
    const duration = lastTs ? timestamp - lastTs : DEFAULT_ANIM_DURATION;
    markerTimestamps.set(id, timestamp);

    const from = marker.getLngLat();
    const startTime = performance.now();

    function step(now: number) {
      const t = Math.min((now - startTime) / duration, 1);
      marker.setLngLat([
        from.lng + (toLng - from.lng) * t,
        from.lat + (toLat - from.lat) * t,
      ]);
      if (t < 1) {
        markerAnimations.set(id, requestAnimationFrame(step));
      } else {
        markerAnimations.delete(id);
      }
    }

    markerAnimations.set(id, requestAnimationFrame(step));
    marker.setRotation(bearing);
  }

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
      if (!gps.last_coordinate) return;
      const { longitude, latitude } = gps.last_coordinate as Required<typeof gps.last_coordinate>;
      const popup = new mapboxgl.Popup({ offset: 25 }).setHTML(
        `<strong>${gps.sn}</strong><br/>${gps.vehicle?.plate_number}`,
      );
      const marker = new mapboxgl.Marker({
        element: makeMarkerEl("20px", "40px", (gps.id ?? 0) % vehicleIcons.length, () => props.onSelectGps(gps.id!)),
      })
        .setLngLat([longitude, latitude])
        .setPopup(popup)
        .addTo(map);
      markerInstances.set(gps.id!, marker);
    });
  });

  // Move marker and rotate to face heading on WS update
  createEffect(() => {
    const update = props.liveUpdate;
    if (!update || !mapReady()) return;

    let marker = markerInstances.get(update.gps_id);

    if (!marker) {
      const gps = props.markers?.find((g) => g.id === update.gps_id);
      if (!gps) return;
      const popup = new mapboxgl.Popup({ offset: 25 }).setHTML(
        `<strong>${gps.sn}</strong><br/>${gps.vehicle?.plate_number}`,
      );
      marker = new mapboxgl.Marker({
        element: makeMarkerEl("28px", "56px", (gps.id ?? 0) % vehicleIcons.length, () => props.onSelectGps(gps.id!)),
      })
        .setLngLat([update.longitude, update.latitude])
        .setPopup(popup)
        .addTo(map);
      markerInstances.set(update.gps_id, marker);
    }

    animateMarker(marker, update.gps_id, update.longitude, update.latitude, update.bearing, update.timestamp);
  });

  // Draw route when history coordinates change
  createEffect(() => {
    if (!mapReady()) return;
    const coords = props.historyCoordinates ?? [];

    if (!map.getSource(HISTORY_SOURCE)) {
      map.addSource(HISTORY_SOURCE, {
        type: "geojson",
        data: { type: "Feature", properties: {}, geometry: { type: "LineString", coordinates: [] } },
      });
      map.addLayer({
        id: HISTORY_LAYER,
        type: "line",
        source: HISTORY_SOURCE,
        layout: { "line-join": "round", "line-cap": "round" },
        paint: { "line-color": "#3b82f6", "line-width": 3, "line-opacity": 0.8 },
      });
    }

    (map.getSource(HISTORY_SOURCE) as mapboxgl.GeoJSONSource).setData({
      type: "Feature",
      properties: {},
      geometry: { type: "LineString", coordinates: coords },
    });
  });

  onCleanup(() => {
    markerAnimations.forEach((id) => cancelAnimationFrame(id));
    markerInstances.forEach((m) => m.remove());
    if (map) map.remove();
  });

  return <div ref={mapContainer} style={{ width: "100%", height: "100%" }} />;
}
