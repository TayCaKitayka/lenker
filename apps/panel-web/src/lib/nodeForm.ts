export interface NodeBootstrapFormState {
  name: string;
  region: string;
  countryCode: string;
  hostname: string;
  expiresInMinutes: string;
}

export type NodeStatus = "pending" | "active" | "unhealthy" | "drained" | "disabled";
export type NodeDrainState = "active" | "draining" | "drained";
export type ConfigRevisionStatus = "pending" | "applied" | "failed" | "rolled_back";

interface NodeLifecycleState {
  status: NodeStatus;
  drain_state: NodeDrainState;
}

interface CreateNodeBootstrapTokenInput {
  name?: string;
  region?: string;
  country_code?: string;
  hostname?: string;
  expires_in_minutes?: number;
}

export function emptyNodeBootstrapForm(): NodeBootstrapFormState {
  return {
    name: "",
    region: "",
    countryCode: "",
    hostname: "",
    expiresInMinutes: "30",
  };
}

export function validateNodeBootstrapForm(form: NodeBootstrapFormState): string | null {
  const expiresInMinutes = parsePositiveInteger(form.expiresInMinutes);

  if (!expiresInMinutes) {
    return "Expiry must be a positive integer.";
  }
  if (expiresInMinutes > 10080) {
    return "Expiry must be 10080 minutes or less.";
  }
  return null;
}

export function buildCreateNodeBootstrapTokenInput(form: NodeBootstrapFormState): CreateNodeBootstrapTokenInput {
  const input: CreateNodeBootstrapTokenInput = {
    expires_in_minutes: parsePositiveInteger(form.expiresInMinutes) ?? 30,
  };

  const name = form.name.trim();
  const region = form.region.trim();
  const countryCode = form.countryCode.trim().toUpperCase();
  const hostname = form.hostname.trim();

  if (name) {
    input.name = name;
  }
  if (region) {
    input.region = region;
  }
  if (countryCode) {
    input.country_code = countryCode;
  }
  if (hostname) {
    input.hostname = hostname;
  }

  return input;
}

export function formatNodeTimestamp(value?: string | null): string {
  if (!value) {
    return "-";
  }

  const parsedValue = new Date(value);
  if (Number.isNaN(parsedValue.getTime())) {
    return "-";
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(parsedValue);
}

export function formatRuntimeEventType(value?: string | null): string {
  const normalizedValue = value?.trim();
  if (!normalizedValue) {
    return "Runtime event";
  }

  return normalizedValue
    .split("_")
    .filter(Boolean)
    .map((part) => part[0].toUpperCase() + part.slice(1))
    .join(" ");
}

export function canDrain(node: NodeLifecycleState): boolean {
  return node.status !== "disabled" && node.drain_state !== "draining";
}

export function canUndrain(node: NodeLifecycleState): boolean {
  return node.status !== "disabled" && node.drain_state !== "active";
}

export function canDisable(node: Pick<NodeLifecycleState, "status">): boolean {
  return node.status !== "disabled";
}

export function canEnable(node: Pick<NodeLifecycleState, "status">): boolean {
  return node.status === "disabled";
}

export function nodeStatusClass(status: NodeStatus): string {
  return `status-${status}`;
}

export function nodeDrainClass(drainState: NodeDrainState): string {
  return drainState === "active" ? "status-active" : "status-draining";
}

export function configRevisionStatusClass(status: ConfigRevisionStatus): string {
  if (status === "applied") {
    return "status-active";
  }
  if (status === "failed") {
    return "status-disabled";
  }
  if (status === "rolled_back") {
    return "status-archived";
  }
  return "status-pending";
}

export function formatConfigRevisionBundle(bundle: unknown): string {
  if (bundle === null || bundle === undefined) {
    return "-";
  }

  if (typeof bundle === "string") {
    return bundle;
  }

  try {
    return JSON.stringify(bundle, null, 2);
  } catch {
    return String(bundle);
  }
}

function parsePositiveInteger(value: string): number | null {
  const normalizedValue = value.trim();
  if (!/^[1-9]\d*$/.test(normalizedValue)) {
    return null;
  }

  const parsedValue = Number(normalizedValue);
  return Number.isSafeInteger(parsedValue) ? parsedValue : null;
}
