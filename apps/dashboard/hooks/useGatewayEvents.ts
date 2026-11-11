"use client";

import { useEffect, useState, useMemo } from "react";
import type { GatewayEventType } from "@/types/gatewayEvent";

export interface GatewayEvent {
  type: GatewayEventType;
  satellite_id: string;
  timestamp: string;
  data?: {
    name?: string;
    battery_percent?: number;
    sequence_number?: number;
    cycles_missed?: number;
    healthy?: boolean;
    accepted?: boolean;
    detail?: string;
    msg?: string;
  }
}

const MAX_EVENTS = 1000;

export function useGatewayEvents(){
  const [events, setEvents] = useState<GatewayEvent[]>([]);
  const [connected, setConnected] = useState(false);

  const telemetryEvents = useMemo(
    () => events.filter((event) => event.type === "satellite.telemetry"
  ),[events])

  const diagnosticEvents = useMemo(() => {
    return events.filter((event) =>
    [
      "satellite.inactive",
      "satellite.ping_sent",
      "satellite.ping_response"
    ].includes(event.type))
  },[events])

  const recoveryEvents = useMemo(() => {
    return events.filter((event) => [
      "satellite.recovery_requested",
      "satellite.command_acknowledged",
      "satellite.online",
    ].includes(event.type))
  },[events])

  useEffect(() => {
    let reconnectTimeout: ReturnType<typeof setTimeout>;
    let socket: WebSocket;
    let stopped = false;

    const connect = () => {
      socket = new WebSocket("ws://localhost:18080/ws");

      socket.onopen = () => {
        setConnected(true)
      };

      socket.onmessage = (message) => {
        try {
          const event = JSON.parse(message.data) as GatewayEvent;

          setEvents((prev) => [
            event,
            ...prev,
          ].slice(0,MAX_EVENTS))
        } catch (error) {
          console.error("Invalid gateway events:", error);
        }
      };

      socket.onclose = () => {
        setConnected(false);

        if (!stopped) {
          reconnectTimeout = setTimeout(connect, 2000);
        }
      };

      socket.onerror = () => {
        socket.close();
      }
    };

    connect();

    return () => {
      stopped = true;
      clearTimeout(reconnectTimeout);
      socket.close();
    }
  },[])

  return {events, telemetryEvents, diagnosticEvents, recoveryEvents, connected};
}