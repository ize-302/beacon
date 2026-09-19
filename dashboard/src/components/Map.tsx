import { useEffect, useRef, useState } from "react";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import type { VehicleResponse } from "~/client/api";
import type { WsCoordinate } from "~/types";
import type { Feature, FeatureCollection, Point } from "geojson";
import policeCarUrl from "~/components/vehicles/police-car.svg?url";

const DEFAULT_ANIM_DURATION = 4000;
const HISTORY_SOURCE = "vehicle-history";
const HISTORY_LAYER = "vehicle-history-line";

const VEHICLES_SOURCE = "vehicles";
const CLUSTER_LAYER = "vehicle-clusters";
const CLUSTER_COUNT_LAYER = "vehicle-cluster-count";
const VEHICLE_POINTS_LAYER = "vehicle-points";
const VEHICLE_ICON = "vehicle-icon";

mapboxgl.accessToken = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;

type VehiclePosition = {
  // currently rendered (possibly mid-interpolation) coordinate
  lng: number;
  lat: number;
  // interpolation endpoints
  fromLng: number;
  fromLat: number;
  toLng: number;
  toLat: number;
  bearing: number;
  startTime: number;
  duration: number;
  // last WS point's own timestamp, used to size the next animation so it
  // matches how long the vehicle actually took between ticks
  lastTimestamp?: number;
  plateNumber: string;
};

export default function DeclarativeMap({
  markers,
  liveUpdates,
  onSelectVehicle,
  historyCoordinates,
}: {
  markers: VehicleResponse[];
  liveUpdates: WsCoordinate[] | null;
  onSelectVehicle: (id: number) => void;
  historyCoordinates: [number, number][] | null;
}) {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<mapboxgl.Map | null>(null);
  const [mapReady, setMapReady] = useState(false);
  const [layersReady, setLayersReady] = useState(false);

  // Always call the latest onSelectVehicle from mapbox's event handlers,
  // which are wired up once in the mount effect below and would otherwise
  // close over a stale callback from the first render.
  const onSelectVehicleRef = useRef(onSelectVehicle);
  useEffect(() => {
    onSelectVehicleRef.current = onSelectVehicle;
  }, [onSelectVehicle]);

  // Same reasoning: the live-update effect only wants to re-run when a new
  // frame arrives, not whenever the marker list refetches, but still needs
  // the current list to fill in plate/device details for a vehicle it
  // hasn't seen yet.
  const markersRef = useRef(markers);
  useEffect(() => {
    markersRef.current = markers;
  }, [markers]);

  // One shared position per vehicle, rendered every frame into a single
  // GeoJSON source. Thousands of individual DOM markers each running their
  // own rAF loop is what actually falls over at fleet scale — a single
  // WebGL-rendered layer, updated by one shared loop, handles it fine.
  const positionsRef = useRef(new Map<number, VehiclePosition>());
  const rafIdRef = useRef<number | null>(null);

  function syncSource() {
    const map = mapRef.current;
    const source = map?.getSource(VEHICLES_SOURCE) as
      | mapboxgl.GeoJSONSource
      | undefined;
    if (!source) return;

    const features: Feature<Point>[] = [];
    positionsRef.current.forEach((p, id) => {
      features.push({
        type: "Feature",
        properties: {
          id,
          plate_number: p.plateNumber,
          bearing: p.bearing,
        },
        geometry: { type: "Point", coordinates: [p.lng, p.lat] },
      });
    });
    const collection: FeatureCollection<Point> = {
      type: "FeatureCollection",
      features,
    };
    source.setData(collection);
  }

  // setData() on a clustered source rebuilds mapbox's cluster index from
  // scratch each call. Pushing at full rAF rate (up to 60/s) against a large
  // fleet can make each rebuild outrun the next call, so newer calls keep
  // preempting older ones before any finishes — a zoom then never gets a
  // completed index to pull new tiles from, and everything vanishes.
  // Interpolation still runs every frame for smoothness; only the push to
  // the map is throttled.
  const SYNC_INTERVAL_MS = 100;
  const lastSyncTimeRef = useRef(0);

  function startAnimationLoop() {
    if (rafIdRef.current !== null) return;

    const step = () => {
      const now = performance.now();
      let stillAnimating = false;

      positionsRef.current.forEach((p) => {
        const t = Math.min((now - p.startTime) / p.duration, 1);
        p.lng = p.fromLng + (p.toLng - p.fromLng) * t;
        p.lat = p.fromLat + (p.toLat - p.fromLat) * t;
        if (t < 1) stillAnimating = true;
      });

      // Always push on the final frame so the resting position is exact,
      // not wherever the throttle last landed.
      if (!stillAnimating || now - lastSyncTimeRef.current >= SYNC_INTERVAL_MS) {
        syncSource();
        lastSyncTimeRef.current = now;
      }

      rafIdRef.current = stillAnimating ? requestAnimationFrame(step) : null;
    };

    rafIdRef.current = requestAnimationFrame(step);
  }

  function popupHtml(plateNumber: string) {
    return `<p style="font-weight:600;font-size:13px;margin:0">${plateNumber}</p>`;
  }

  function initVehicleLayers() {
    const map = mapRef.current;
    if (!map) return;

    map.addSource(VEHICLES_SOURCE, {
      type: "geojson",
      data: { type: "FeatureCollection", features: [] },
      cluster: true,
      clusterMaxZoom: 14,
      clusterRadius: 50,
    });

    map.addLayer({
      id: CLUSTER_LAYER,
      type: "circle",
      source: VEHICLES_SOURCE,
      filter: ["has", "point_count"],
      paint: {
        "circle-color": [
          "step",
          ["get", "point_count"],
          "#51bbd6",
          25,
          "#f1c40f",
          100,
          "#e74c3c",
        ],
        "circle-radius": ["step", ["get", "point_count"], 16, 25, 22, 100, 28],
        "circle-stroke-width": 2,
        "circle-stroke-color": "#fff",
      },
    });

    map.addLayer({
      id: CLUSTER_COUNT_LAYER,
      type: "symbol",
      source: VEHICLES_SOURCE,
      filter: ["has", "point_count"],
      layout: {
        "text-field": ["get", "point_count_abbreviated"],
        "text-font": ["DIN Offc Pro Medium", "Arial Unicode MS Bold"],
        "text-size": 13,
      },
      paint: { "text-color": "#1f2937" },
    });

    map.addLayer({
      id: VEHICLE_POINTS_LAYER,
      type: "symbol",
      source: VEHICLES_SOURCE,
      filter: ["!", ["has", "point_count"]],
      layout: {
        "icon-image": VEHICLE_ICON,
        "icon-size": 0.22,
        "icon-rotate": ["get", "bearing"],
        "icon-rotation-alignment": "map",
        "icon-allow-overlap": true,
        "icon-ignore-placement": true,
      },
    });

    // Clicking a cluster zooms to the point where it splits apart, same
    // pattern as mapbox's own clustering example.
    map.on("click", CLUSTER_LAYER, (e) => {
      const features = map.queryRenderedFeatures(e.point, {
        layers: [CLUSTER_LAYER],
      });
      const clusterId = features[0]?.properties?.cluster_id;
      if (clusterId === undefined) return;
      const source = map.getSource(VEHICLES_SOURCE) as mapboxgl.GeoJSONSource;
      source.getClusterExpansionZoom(clusterId, (err, zoom) => {
        if (err || zoom == null) return;
        map.easeTo({
          center: (features[0].geometry as Point).coordinates as [
            number,
            number,
          ],
          zoom,
        });
      });
    });

    map.on("click", VEHICLE_POINTS_LAYER, (e) => {
      const f = e.features?.[0];
      if (!f) return;
      const p = f.properties as {
        id: number;
        plate_number: string;
      };
      const coords = (f.geometry as Point).coordinates.slice() as [
        number,
        number,
      ];
      new mapboxgl.Popup({ offset: 25 })
        .setLngLat(coords)
        .setHTML(popupHtml(p.plate_number))
        .addTo(map);
      onSelectVehicleRef.current(p.id);
    });

    for (const layer of [CLUSTER_LAYER, VEHICLE_POINTS_LAYER]) {
      map.on("mouseenter", layer, () => {
        map.getCanvas().style.cursor = "pointer";
      });
      map.on("mouseleave", layer, () => {
        map.getCanvas().style.cursor = "";
      });
    }

    setLayersReady(true);
  }

  useEffect(() => {
    if (!mapContainerRef.current) return;

    const map = new mapboxgl.Map({
      container: mapContainerRef.current,
      style: "mapbox://styles/mapbox/streets-v12",
      center: [3.37936, 6.5103],
      zoom: 8,
      pitchWithRotate: false,
      maxPitch: 0,
    });
    mapRef.current = map;
    map.addControl(new mapboxgl.NavigationControl(), "top-right");
    map.on("load", () => {
      setMapReady(true);
      const img = new Image();
      img.onload = () => {
        map.addImage(VEHICLE_ICON, img);
        initVehicleLayers();
      };
      img.src = policeCarUrl;
    });

    return () => {
      if (rafIdRef.current !== null) cancelAnimationFrame(rafIdRef.current);
      map.remove();
      mapRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Seed positions from the REST vehicle list. Only fills in vehicles we
  // don't already know about, so a refetch never snaps an actively-animating
  // vehicle back to a stale coordinate.
  useEffect(() => {
    if (!layersReady || !markers?.length) return;

    let changed = false;
    markers.forEach((v) => {
      if (v.id == null) return;

      const existing = positionsRef.current.get(v.id);
      if (existing) {
        existing.plateNumber = v.plate_number;
        return;
      }

      if (!v.last_coordinate) return;
      const { longitude, latitude } = v.last_coordinate as Required<
        typeof v.last_coordinate
      >;
      positionsRef.current.set(v.id, {
        lng: longitude,
        lat: latitude,
        fromLng: longitude,
        fromLat: latitude,
        toLng: longitude,
        toLat: latitude,
        bearing: 0,
        startTime: 0,
        duration: DEFAULT_ANIM_DURATION,
        plateNumber: v.plate_number,
      });
      changed = true;
    });

    if (changed) syncSource();
  }, [layersReady, markers]);

  // A frame carries every position recorded in the same write, so animate
  // them all toward their new coordinate.
  useEffect(() => {
    const frame = liveUpdates;
    if (!frame?.length || !layersReady) return;

    for (const update of frame) {
      let p = positionsRef.current.get(update.vehicle_id);
      if (!p) {
        const v = markersRef.current?.find((m) => m.id === update.vehicle_id);
        p = {
          lng: update.longitude,
          lat: update.latitude,
          fromLng: update.longitude,
          fromLat: update.latitude,
          toLng: update.longitude,
          toLat: update.latitude,
          bearing: update.bearing,
          startTime: 0,
          duration: DEFAULT_ANIM_DURATION,
          plateNumber: v?.plate_number ?? "",
        };
        positionsRef.current.set(update.vehicle_id, p);
      }

      const duration = p.lastTimestamp
        ? update.timestamp - p.lastTimestamp
        : DEFAULT_ANIM_DURATION;
      p.lastTimestamp = update.timestamp;

      p.fromLng = p.lng;
      p.fromLat = p.lat;
      p.toLng = update.longitude;
      p.toLat = update.latitude;
      p.bearing = update.bearing;
      p.startTime = performance.now();
      p.duration = Math.max(duration, 16);
    }

    startAnimationLoop();
  }, [liveUpdates, layersReady]);

  // Draw route when history coordinates change
  useEffect(() => {
    const map = mapRef.current;
    if (!mapReady || !map) return;
    const coords = historyCoordinates ?? [];

    if (!map.getSource(HISTORY_SOURCE)) {
      map.addSource(HISTORY_SOURCE, {
        type: "geojson",
        data: {
          type: "Feature",
          properties: {},
          geometry: { type: "LineString", coordinates: [] },
        },
      });
      map.addLayer({
        id: HISTORY_LAYER,
        type: "line",
        source: HISTORY_SOURCE,
        layout: { "line-join": "round", "line-cap": "round" },
        paint: {
          "line-color": "#3b82f6",
          "line-width": 3,
          "line-opacity": 0.8,
        },
      });
    }

    (map.getSource(HISTORY_SOURCE) as mapboxgl.GeoJSONSource).setData({
      type: "Feature",
      properties: {},
      geometry: { type: "LineString", coordinates: coords },
    });
  }, [mapReady, historyCoordinates]);

  return (
    <div ref={mapContainerRef} style={{ width: "100%", height: "100%" }} />
  );
}
