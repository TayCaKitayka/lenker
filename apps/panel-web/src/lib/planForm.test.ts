import {
  buildCreatePlanInput,
  buildUpdatePlanInput,
  emptyPlanForm,
  planToForm,
  validatePlanForm,
} from "./planForm";

function assert(condition: boolean, message: string): void {
  if (!condition) {
    throw new Error(message);
  }
}

function runTests(): void {
  validatesRequiredFieldsAndPositiveIntegers();
  buildsCreateInputWithoutTrafficLimit();
  buildsCreateInputWithTrafficLimit();
  buildsUpdateInputWithClearTrafficLimit();
  mapsPlanToForm();
}

function validatesRequiredFieldsAndPositiveIntegers(): void {
  const form = emptyPlanForm();

  assert(validatePlanForm({ ...form, name: "" }) === "Plan name is required.", "expected missing name validation");
  assert(
    validatePlanForm({ ...form, name: "Monthly", durationDays: "0" }) ===
      "Duration days must be a positive integer.",
    "expected duration validation",
  );
  assert(
    validatePlanForm({ ...form, name: "Monthly", deviceLimit: "-1" }) ===
      "Device limit must be a positive integer.",
    "expected device limit validation",
  );
  assert(
    validatePlanForm({ ...form, name: "Monthly", hasTrafficLimit: true, trafficLimitBytes: "" }) ===
      "Traffic limit must be a positive integer when enabled.",
    "expected traffic limit validation",
  );
  assert(validatePlanForm({ ...form, name: "Monthly" }) === null, "expected valid plan form");
}

function buildsCreateInputWithoutTrafficLimit(): void {
  const input = buildCreatePlanInput({
    name: " Monthly ",
    durationDays: "30",
    deviceLimit: "3",
    hasTrafficLimit: false,
    trafficLimitBytes: "",
  });

  assert(input.name === "Monthly", "expected trimmed plan name");
  assert(input.duration_days === 30, "expected duration days");
  assert(input.device_limit === 3, "expected device limit");
  assert(!("traffic_limit_bytes" in input), "expected omitted traffic limit");
}

function buildsCreateInputWithTrafficLimit(): void {
  const input = buildCreatePlanInput({
    name: "Monthly",
    durationDays: "30",
    deviceLimit: "3",
    hasTrafficLimit: true,
    trafficLimitBytes: "1073741824",
  });

  assert(input.traffic_limit_bytes === 1073741824, "expected traffic limit");
}

function buildsUpdateInputWithClearTrafficLimit(): void {
  const input = buildUpdatePlanInput({
    name: "Monthly",
    durationDays: "30",
    deviceLimit: "3",
    hasTrafficLimit: false,
    trafficLimitBytes: "",
  });

  assert(input.clear_traffic_limit === true, "expected clear traffic limit flag");
  assert(!("traffic_limit_bytes" in input), "expected no traffic limit value when clearing");
}

function mapsPlanToForm(): void {
  const form = planToForm({
    name: "Annual",
    duration_days: 365,
    traffic_limit_bytes: 1099511627776,
    device_limit: 5,
  });

  assert(form.name === "Annual", "expected plan name");
  assert(form.durationDays === "365", "expected duration days");
  assert(form.deviceLimit === "5", "expected device limit");
  assert(form.hasTrafficLimit, "expected enabled traffic limit");
  assert(form.trafficLimitBytes === "1099511627776", "expected traffic limit bytes");
}

runTests();
