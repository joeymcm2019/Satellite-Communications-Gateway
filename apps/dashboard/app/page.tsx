"use client";

import { useState } from "react";
import { Panel } from "@/components/panel";
import { Satellite } from "@/components/satellite";
import { useSatellites } from "@/hooks/useSatellites";
import { useGatewayEvents } from "@/hooks/useGatewayEvents";


type EventFilter = "all" | "telemetry" | "diagnostic" | "recovery";

export default function Home() {

const {satellites, error} = useSatellites();

const {events, telemetryEvents, diagnosticEvents, recoveryEvents, connected} = useGatewayEvents();

const satellitesOnline = satellites?.filter(sat => sat.status === "online").length ?? 0;

const [eventFilter, setEventFilter] = useState<EventFilter>("all");

const filteredEvents = {
  all: events,
  telemetry: telemetryEvents,
  diagnostic: diagnosticEvents,
  recovery: recoveryEvents,
}[eventFilter];

  return (
    <main>
      <header>
        <div>
          <p className="eyebrow">Mission control / local simulation</p>
          <h1>Satellite Communications Gateway</h1>
        </div>
      </header>
      <div className="grid">
        <Panel title="Telemetry & event log" label="03 / Activity" className="wide">
          <p className="hint" role="status">
            {connected ? "Live event stream connected" : "Connecting to gateway..."}
          </p>

          <div className="event-filters" role="group" aria-label="Filter gateway events">
            {(["all", "telemetry", "diagnostic", "recovery"] as const).map((filter) => (
              <button
                key={filter}
                type="button"
                aria-pressed={eventFilter === filter}
                onClick={() => setEventFilter(filter)}
                className={`${eventFilter === filter && "selected"}`}
              >
                {filter.charAt(0).toUpperCase() + filter.slice(1)}
              </button>
            ))}
          </div>

          {!filteredEvents ? (
            <div className="empty">No {eventFilter !== "all" && eventFilter} events received</div>
          ) : (
            <div className="event-list" role="log" aria-label="Gateway activity">
              {filteredEvents.map((event, index) => (
                <div className="log" key={`${event.timestamp}-${event.satellite_id}-${index}`}>
                  <span className="log-tag">
                    {event.type.replace("satellite.", "").toUpperCase()}
                  </span>

                  <p style={{display: "flex"}}>
                    <strong style={{width: "59px", display: "flex"}}>{event.satellite_id}</strong>
                    {event.type === "satellite.telemetry" && (
                      <>
                        <span style={{width: "130px"}}>
                          {" "}— Battery: {(event.data?.battery_percent || 0).toFixed(2)}%
                        </span>
                        <span style={{minWidth: "130px"}}>
                          {" "}· Sequence: {event.data?.sequence_number}
                        </span>
                      </>
                    )}

                    <small style={{paddingLeft: "5px"}}>
                      {" "}{new Date(event.timestamp).toLocaleTimeString()}
                    </small>
                  </p>
                </div>
              ))}
            </div>
          )}
        </Panel>
        <Panel title="Satellite fleet" label="02 / Registry">
          <p className="hint">Illustrative satellite identifiers</p>
          <ul className="satellites">
            {satellites?.length > 0 && [...satellites].sort((a, b) => a.name.localeCompare(b.name)).map((sat) =>
              <Satellite key={sat.id} satellite={sat} />
            )}
          </ul>
        </Panel>
        <Panel title="Network overview" label="03 / Network" className="wide">
          <div className="metrics">
            {[["Gateway", !error ? "Connected" : "Not connected"], ["Satellites online", satellitesOnline], ["Active Outages", satellites.length - satellitesOnline],].map(([label, value]) =>
              <div className="metric" key={label}>
                <span>{label}</span><strong>{value}</strong>
              </div>
            )}
          </div>
          {/* <div className="network">
            <div className="orbit" aria-hidden="true">
              <span>◎</span>
            </div>
            <p>Awaiting first contact</p>
            <small>The future gRPC → gateway → WebSocket link will appear here.</small>
          </div> */}
        </Panel>
        <Panel title="Communication controls" label="04 / Uplink">
          <p className="hint">
            Command delivery is planned for a later milestone.
          </p>
          <fieldset disabled>
            <label htmlFor="satellite">
              Target satellite
            </label>
            <select id="satellite">
              <option>Select a satellite</option>
            </select>
            <label htmlFor="command">Command</label>
            <select id="command">
              <option>Request telemetry</option>
            </select>
            <button type="button">
              Send simulated command <span aria-hidden="true">↗</span>
            </button>
          </fieldset>
        </Panel>
      </div>
      <footer>
        <span>Independent educational project</span>
        <span>Simulated infrastructure only · v1.0</span>
      </footer>
    </main>
  );
}
