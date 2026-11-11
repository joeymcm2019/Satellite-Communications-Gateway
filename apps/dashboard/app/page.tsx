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

console.log(events)

  return (
    <main>
      <header>
        <div>
          <p className="eyebrow">Mission control / local simulation</p>
          <h1>Satellite Communications Gateway</h1>
        </div>
      </header>
      <div className="grid">
        <Panel title="Telemetry & event log" label="03 / Activity" className="wide minH">
          <p className="hint" role="status">
            {connected ? "Live event stream connected" : "Connecting to gateway..."}
          </p>

          {filteredEvents &&
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
          }

          {!filteredEvents || filteredEvents.length === 0 ? (
            <div className="empty">No {eventFilter !== "all" && eventFilter} events received</div>
          ) : (
            <div className="event-list" role="log" aria-label="Gateway activity">
              {filteredEvents.map((event, index) => (
                <div className="log" key={`${event.timestamp}-${event.satellite_id}-${index}`}>
                  <span className={`log-tag event-${event.type.replace("satellite.", "")}`}>
                    {event.type.replace("satellite.", "").replace("_", " ").toUpperCase()}
                  </span>

                  <p style={{display: "flex" , justifyContent: "space-between", width: "100%", paddingRight: "20px"}}>

                    {event.type === "satellite.telemetry" && (
                      <span>
                        <strong style={{width: "59px"}}>
                          {event.satellite_id}
                        </strong>
                        <span style={{width: "130px"}}>
                          {" "}— Battery: {(event.data?.battery_percent || 0).toFixed(2)}%
                        </span>
                        <span style={{minWidth: "130px"}}>
                          {" "}· Sequence: {event.data?.sequence_number}
                        </span>
                      </span>
                    )}

                    {event.type === "satellite.inactive" && (
                      <span>
                        <strong style={{width: "59px"}}>
                            {event.satellite_id}
                        </strong>
                        <span style={{minWidth: "260px"}}>
                          {" "}— Inactive & Missed {event.data?.cycles_missed} Cycles
                        </span>
                      </span>
                    )}

                    {event.type === "satellite.ping_sent" && (
                      <>
                      <span style={{minWidth: "319px"}}>
                        Ping Sent to
                        <strong>
                          {" "}{event.satellite_id}
                        </strong>
                      </span>
                      </>
                    )}

                    {event.type === "satellite.ping_response" && (
                      <>
                      <span style={{minWidth: "319px"}}>
                        Ping response from
                        <strong>
                          {" "}{event.satellite_id}
                        </strong>
                        : healthy=
                        <span className={`healthy-${event.data?.healthy}`}>
                          {event.data?.healthy ? "true" : "false"}
                        </span>
                      </span>
                      </>
                    )}

                    {event.type === "satellite.recovery_requested" && (
                      <>
                       <span style={{minWidth: "319px"}}>
                        Attempting recovery for satellite
                        <strong>
                          {" "}{event.satellite_id}
                        </strong>
                      </span>
                      </>
                    )}

                    {event.type === "satellite.command_acknowledged" && (
                       <span style={{minWidth: "319px"}}>
                        <strong>
                          {" "}{event.satellite_id}
                        </strong>
                        <span className="small">{" "}— Command_Accepted=</span>
                        <span className={`small healthy-${event.data?.accepted}`}>
                          {event.data?.accepted ? "true" : "false"}
                        </span>
                        <span className={"small"}>{" "}Details=</span>
                        <span className={"small"}>
                          {`"${event.data?.detail}"`}
                        </span>
                      </span>
                    )}

                    {event.type === "satellite.online" && (
                      <span style={{minWidth: "319px"}}>
                        <strong style={{width: "59px"}}>
                            {event.satellite_id}
                        </strong>
                        <span style={{minWidth: "260px"}} className="small">
                          {" "}— Online. {event.data?.msg}
                        </span>
                      </span>
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
