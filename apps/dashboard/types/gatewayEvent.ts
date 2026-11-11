export const GatewayEventType = {
  Telemetry: "satellite.telemetry",
  Inactive: "satellite.inactive",
  PingSent: "satellite.ping_sent",
  PingResponse: "satellite.ping_response",
  RecoveryRequested: "satellite.recovery_requested",
  CommandAcknowledged: "satellite.command_acknowledged",
  Online: "satellite.online",
} as const;

export type GatewayEventType =
  (typeof GatewayEventType)[keyof typeof GatewayEventType];