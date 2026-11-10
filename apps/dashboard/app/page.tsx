"use client";

import { Panel } from "@/components/panel";
import { Satellite } from "@/components/satellite";
import {useSatellites} from "@/hooks/useSatellites";

export default function Home() {

const {satellites, loading, error} = useSatellites();

const satellitesOnline = satellites?.filter(sat => sat.status === "online").length ?? 0;

  return (
    <main>
      <header><div><p className="eyebrow">Mission control / local simulation</p>
        <h1>Satellite Communications Gateway</h1>
        <p className="subtitle">A foundation for exploring communication beyond the horizon.</p></div>

      </header>
      <div className="grid">
        <Panel title="Network overview" label="01 / Network" className="wide">
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
        <Panel title="Satellite fleet" label="02 / Registry">
          <p className="hint">Illustrative satellite identifiers</p>
          <ul className="satellites">
            {satellites?.length > 0 && [...satellites].sort((a, b) => a.name.localeCompare(b.name)).map((sat) =>
              <Satellite key={sat.id} satellite={sat} />
            )}
          </ul>
        </Panel>
        <Panel title="Telemetry & event log" label="03 / Activity" className="wide">
          <div className="log">
            <span className="log-tag">SYSTEM</span>
            <p>Dashboard scaffold ready.<br/>
            <small>Telemetry and gateway events will appear here when connected.</small>
            </p>
          </div>
          <div className="empty">No events received</div>
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
        <span>Simulated infrastructure only · v0.1</span>
      </footer>
    </main>
  );
}
