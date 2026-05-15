export interface PlanFormState {
  name: string;
  durationDays: string;
  deviceLimit: string;
  hasTrafficLimit: boolean;
  trafficLimitBytes: string;
}

export interface PlanFormSource {
  name: string;
  duration_days: number;
  traffic_limit_bytes: number | null;
  device_limit: number;
}

export interface CreatePlanFormInput {
  name: string;
  duration_days: number;
  traffic_limit_bytes?: number | null;
  device_limit: number;
}

export interface UpdatePlanFormInput {
  name: string;
  duration_days: number;
  traffic_limit_bytes?: number | null;
  clear_traffic_limit?: boolean;
  device_limit: number;
}

export function emptyPlanForm(): PlanFormState {
  return {
    name: "",
    durationDays: "30",
    deviceLimit: "1",
    hasTrafficLimit: false,
    trafficLimitBytes: "",
  };
}

export function planToForm(plan: PlanFormSource): PlanFormState {
  return {
    name: plan.name,
    durationDays: String(plan.duration_days),
    deviceLimit: String(plan.device_limit),
    hasTrafficLimit: plan.traffic_limit_bytes !== null,
    trafficLimitBytes: plan.traffic_limit_bytes === null ? "" : String(plan.traffic_limit_bytes),
  };
}

export function validatePlanForm(form: PlanFormState): string | null {
  if (!form.name.trim()) {
    return "Plan name is required.";
  }
  if (!parsePositiveInteger(form.durationDays)) {
    return "Duration days must be a positive integer.";
  }
  if (!parsePositiveInteger(form.deviceLimit)) {
    return "Device limit must be a positive integer.";
  }
  if (form.hasTrafficLimit && !parsePositiveInteger(form.trafficLimitBytes)) {
    return "Traffic limit must be a positive integer when enabled.";
  }
  return null;
}

export function buildCreatePlanInput(form: PlanFormState): CreatePlanFormInput {
  const input: CreatePlanFormInput = {
    name: form.name.trim(),
    duration_days: parsePositiveInteger(form.durationDays) ?? 0,
    device_limit: parsePositiveInteger(form.deviceLimit) ?? 0,
  };

  if (form.hasTrafficLimit) {
    input.traffic_limit_bytes = parsePositiveInteger(form.trafficLimitBytes) ?? 0;
  }

  return input;
}

export function buildUpdatePlanInput(form: PlanFormState): UpdatePlanFormInput {
  const input: UpdatePlanFormInput = {
    name: form.name.trim(),
    duration_days: parsePositiveInteger(form.durationDays) ?? 0,
    device_limit: parsePositiveInteger(form.deviceLimit) ?? 0,
  };

  if (form.hasTrafficLimit) {
    input.traffic_limit_bytes = parsePositiveInteger(form.trafficLimitBytes) ?? 0;
  } else {
    input.clear_traffic_limit = true;
  }

  return input;
}

function parsePositiveInteger(value: string): number | null {
  const normalizedValue = value.trim();
  if (!/^[1-9]\d*$/.test(normalizedValue)) {
    return null;
  }

  const parsedValue = Number(normalizedValue);
  return Number.isSafeInteger(parsedValue) ? parsedValue : null;
}
