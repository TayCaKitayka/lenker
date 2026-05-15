export type SubscriptionStatus = "active" | "expired" | "suspended";

export interface SubscriptionFormState {
  userID: string;
  planID: string;
  status: SubscriptionStatus;
  deviceLimit: string;
  hasTrafficLimit: boolean;
  trafficLimitBytes: string;
  hasPreferredRegion: boolean;
  preferredRegion: string;
  renewDays: string;
}

export interface SubscriptionFormSource {
  user_id: string;
  plan_id: string;
  status: SubscriptionStatus;
  traffic_limit_bytes: number | null;
  device_limit: number;
  preferred_region: string | null;
}

export interface CreateSubscriptionFormInput {
  user_id: string;
  plan_id: string;
  preferred_region?: string | null;
}

export interface UpdateSubscriptionFormInput {
  status: SubscriptionStatus;
  traffic_limit_bytes?: number | null;
  clear_traffic_limit?: boolean;
  device_limit: number;
  preferred_region?: string | null;
  clear_preferred_region?: boolean;
}

export interface RenewSubscriptionFormInput {
  extend_days: number;
}

export function emptySubscriptionForm(): SubscriptionFormState {
  return {
    userID: "",
    planID: "",
    status: "active",
    deviceLimit: "1",
    hasTrafficLimit: false,
    trafficLimitBytes: "",
    hasPreferredRegion: false,
    preferredRegion: "",
    renewDays: "30",
  };
}

export function subscriptionToForm(subscription: SubscriptionFormSource): SubscriptionFormState {
  return {
    userID: subscription.user_id,
    planID: subscription.plan_id,
    status: subscription.status,
    deviceLimit: String(subscription.device_limit),
    hasTrafficLimit: subscription.traffic_limit_bytes !== null,
    trafficLimitBytes: subscription.traffic_limit_bytes === null ? "" : String(subscription.traffic_limit_bytes),
    hasPreferredRegion: subscription.preferred_region !== null && subscription.preferred_region.trim() !== "",
    preferredRegion: subscription.preferred_region ?? "",
    renewDays: "30",
  };
}

export function validateCreateSubscriptionForm(form: SubscriptionFormState): string | null {
  if (!form.userID.trim()) {
    return "User is required.";
  }
  if (!form.planID.trim()) {
    return "Plan is required.";
  }
  return null;
}

export function validateUpdateSubscriptionForm(form: SubscriptionFormState): string | null {
  if (!isKnownStatus(form.status)) {
    return "Status must be active, expired, or suspended.";
  }
  if (!parsePositiveInteger(form.deviceLimit)) {
    return "Device limit must be a positive integer.";
  }
  if (form.hasTrafficLimit && !parsePositiveInteger(form.trafficLimitBytes)) {
    return "Traffic limit must be a positive integer when enabled.";
  }
  return null;
}

export function validateRenewSubscriptionForm(form: SubscriptionFormState): string | null {
  if (!parsePositiveInteger(form.renewDays)) {
    return "Renew days must be a positive integer.";
  }
  return null;
}

export function buildCreateSubscriptionInput(form: SubscriptionFormState): CreateSubscriptionFormInput {
  const input: CreateSubscriptionFormInput = {
    user_id: form.userID.trim(),
    plan_id: form.planID.trim(),
  };

  if (form.hasPreferredRegion) {
    input.preferred_region = form.preferredRegion.trim();
  }

  return input;
}

export function buildUpdateSubscriptionInput(form: SubscriptionFormState): UpdateSubscriptionFormInput {
  const input: UpdateSubscriptionFormInput = {
    status: form.status,
    device_limit: parsePositiveInteger(form.deviceLimit) ?? 0,
  };

  if (form.hasTrafficLimit) {
    input.traffic_limit_bytes = parsePositiveInteger(form.trafficLimitBytes) ?? 0;
  } else {
    input.clear_traffic_limit = true;
  }

  if (form.hasPreferredRegion) {
    input.preferred_region = form.preferredRegion.trim();
  } else {
    input.clear_preferred_region = true;
  }

  return input;
}

export function buildRenewSubscriptionInput(form: SubscriptionFormState): RenewSubscriptionFormInput {
  return {
    extend_days: parsePositiveInteger(form.renewDays) ?? 0,
  };
}

function isKnownStatus(status: string): status is SubscriptionStatus {
  return status === "active" || status === "expired" || status === "suspended";
}

function parsePositiveInteger(value: string): number | null {
  const normalizedValue = value.trim();
  if (!/^[1-9]\d*$/.test(normalizedValue)) {
    return null;
  }

  const parsedValue = Number(normalizedValue);
  return Number.isSafeInteger(parsedValue) ? parsedValue : null;
}
