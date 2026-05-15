import {
  buildCreateSubscriptionInput,
  buildRenewSubscriptionInput,
  buildUpdateSubscriptionInput,
  emptySubscriptionForm,
  subscriptionToForm,
  validateCreateSubscriptionForm,
  validateRenewSubscriptionForm,
  validateUpdateSubscriptionForm,
} from "./subscriptionForm";

function assert(condition: boolean, message: string): void {
  if (!condition) {
    throw new Error(message);
  }
}

function runTests(): void {
  validatesCreateRequirements();
  validatesUpdatePositiveIntegers();
  buildsCreateInputWithPreferredRegion();
  buildsUpdateInputWithClearFlags();
  buildsRenewInput();
  mapsSubscriptionToForm();
}

function validatesCreateRequirements(): void {
  const form = emptySubscriptionForm();

  assert(validateCreateSubscriptionForm(form) === "User is required.", "expected missing user validation");
  assert(validateCreateSubscriptionForm({ ...form, userID: "user-1" }) === "Plan is required.", "expected missing plan validation");
  assert(validateCreateSubscriptionForm({ ...form, userID: "user-1", planID: "plan-1" }) === null, "expected valid create form");
}

function validatesUpdatePositiveIntegers(): void {
  const form = emptySubscriptionForm();

  assert(
    validateUpdateSubscriptionForm({ ...form, deviceLimit: "0" }) === "Device limit must be a positive integer.",
    "expected device limit validation",
  );
  assert(
    validateUpdateSubscriptionForm({ ...form, hasTrafficLimit: true, trafficLimitBytes: "" }) ===
      "Traffic limit must be a positive integer when enabled.",
    "expected traffic limit validation",
  );
  assert(validateUpdateSubscriptionForm({ ...form, deviceLimit: "3" }) === null, "expected valid update form");
}

function buildsCreateInputWithPreferredRegion(): void {
  const input = buildCreateSubscriptionInput({
    ...emptySubscriptionForm(),
    userID: " user-1 ",
    planID: " plan-1 ",
    hasPreferredRegion: true,
    preferredRegion: " eu ",
  });

  assert(input.user_id === "user-1", "expected trimmed user id");
  assert(input.plan_id === "plan-1", "expected trimmed plan id");
  assert(input.preferred_region === "eu", "expected trimmed preferred region");
}

function buildsUpdateInputWithClearFlags(): void {
  const input = buildUpdateSubscriptionInput({
    ...emptySubscriptionForm(),
    status: "suspended",
    deviceLimit: "3",
    hasTrafficLimit: false,
    hasPreferredRegion: false,
  });

  assert(input.status === "suspended", "expected status");
  assert(input.device_limit === 3, "expected device limit");
  assert(input.clear_traffic_limit === true, "expected clear traffic limit");
  assert(input.clear_preferred_region === true, "expected clear preferred region");
}

function buildsRenewInput(): void {
  const input = buildRenewSubscriptionInput({
    ...emptySubscriptionForm(),
    renewDays: "45",
  });

  assert(validateRenewSubscriptionForm({ ...emptySubscriptionForm(), renewDays: "0" }) === "Renew days must be a positive integer.", "expected renew validation");
  assert(input.extend_days === 45, "expected renew days");
}

function mapsSubscriptionToForm(): void {
  const form = subscriptionToForm({
    user_id: "user-1",
    plan_id: "plan-1",
    status: "active",
    traffic_limit_bytes: 1073741824,
    device_limit: 3,
    preferred_region: "eu",
  });

  assert(form.userID === "user-1", "expected user id");
  assert(form.planID === "plan-1", "expected plan id");
  assert(form.hasTrafficLimit, "expected enabled traffic limit");
  assert(form.trafficLimitBytes === "1073741824", "expected traffic limit");
  assert(form.hasPreferredRegion, "expected enabled preferred region");
  assert(form.preferredRegion === "eu", "expected preferred region");
}

runTests();
