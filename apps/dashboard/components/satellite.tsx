"use client";

import { useId, useState } from "react";
import type { Satellite as SatelliteData } from "@/hooks/useSatellites";

export function Satellite({ satellite }: { satellite: SatelliteData }) {
  const [expanded, setExpanded] = useState(false);
  const detailsId = useId();
  const lastSeen = new Date(satellite.last_seen);
  const batteryLevel = satellite.battery_percent <= 20 ? "low" : satellite.battery_percent < 50 ? "medium" : "high";

  return (
    <li className="satellite">
      <button
        type="button"
        className="satellite-toggle"
        aria-expanded={expanded}
        aria-controls={detailsId}
        onClick={() => setExpanded((current) => !current)}
      >
        <span className="sat-icon" aria-hidden="true">◇</span>
        <span className="satellite-identity">
          <strong>{satellite.id}</strong>
          <small>{satellite.name}</small>
        </span>
        <span className={`status status-${satellite.status}`}>
          {satellite.status.charAt(0).toUpperCase() + satellite.status.slice(1)}
        </span>
        <svg className="satellite-chevron" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <path d="m6 15 6-6 6 6" />
        </svg>
      </button>
      <div id={detailsId} className={`satellite-dropdown${expanded ? " is-expanded" : ""}`} aria-hidden={!expanded} inert={!expanded}>
        <div className="satellite-dropdown-inner">
        <dl className="satellite-details">
          <div><dt>Battery {satellite.charging && "(charging)"}</dt><dd className={`battery-${batteryLevel}`}>{satellite.battery_percent}%</dd></div>
          <div><dt>Sequence number</dt><dd>{satellite.sequence_number}</dd></div>
          <div><dt>Last seen</dt><dd>{Number.isNaN(lastSeen.getTime()) ? "Unknown" : <time dateTime={satellite.last_seen}>{lastSeen.toLocaleString()}</time>}</dd></div>
          <div><dt>Expected interval</dt><dd>{satellite.expected_interval_seconds} seconds</dd></div>
        </dl>
        </div>
      </div>
    </li>
  );
}
