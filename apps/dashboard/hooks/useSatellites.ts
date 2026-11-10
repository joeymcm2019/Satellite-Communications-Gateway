"use client";

import {useEffect, useState} from "react";

export interface Satellite {
  id: string;
  name: string
  status: "online" | "inactive";
  battery_percent: number;
  sequence_number: number;
  last_seen: string;
  expected_interval_seconds: number;
  charging: boolean;
}

interface SatellitesResponse {
  satellites: Satellite[];
}

export function useSatellites() {
  const [satellites, setSatellites] = useState<Satellite[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();

    const fetchSatellites = async () => {
      try {
        const response = await fetch("/api/satellites", {
          signal: controller.signal,
          cache: "no-store",
        });

        if (!response.ok){
          throw new Error(`Failed to fetch satellites. Https error. Status: ${response.status}`)
        }

        const data: SatellitesResponse = await response.json();

        if (controller.signal.aborted) return;

        setSatellites(data.satellites);
        setError(null);
      } catch (err) {
        if (controller.signal.aborted) return;

        setError(
          err instanceof Error ? err.message : "Unknown error",
        );
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      }
    }

    fetchSatellites();

    const interval = setInterval(fetchSatellites, 3000);

    return () => {
      controller.abort()
      clearInterval(interval);
    }
  },[])

  return {satellites, loading, error}
}