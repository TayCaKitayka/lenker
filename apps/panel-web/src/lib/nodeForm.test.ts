import {
  buildCreateNodeBootstrapTokenInput,
  canDisable,
  canDrain,
  canEnable,
  canUndrain,
  configRevisionStatusClass,
  emptyNodeBootstrapForm,
  formatConfigRevisionBundle,
  formatNodeTimestamp,
  formatRuntimeEventType,
  validateNodeBootstrapForm,
} from "./nodeForm";

function assert(condition: boolean, message: string): void {
  if (!condition) {
    throw new Error(message);
  }
}

function runTests(): void {
  validatesExpiry();
  buildsTrimmedBootstrapInput();
  formatsTimestampsSafely();
  formatsRuntimeEventTypes();
  checksLifecycleActions();
  formatsConfigRevisionDisplay();
}

function validatesExpiry(): void {
  assert(validateNodeBootstrapForm({ ...emptyNodeBootstrapForm(), expiresInMinutes: "0" }) === "Expiry must be a positive integer.", "expected positive expiry validation");
  assert(validateNodeBootstrapForm({ ...emptyNodeBootstrapForm(), expiresInMinutes: "10081" }) === "Expiry must be 10080 minutes or less.", "expected max expiry validation");
  assert(validateNodeBootstrapForm({ ...emptyNodeBootstrapForm(), expiresInMinutes: "30" }) === null, "expected valid expiry");
}

function buildsTrimmedBootstrapInput(): void {
  const input = buildCreateNodeBootstrapTokenInput({
    ...emptyNodeBootstrapForm(),
    name: " finland-1 ",
    region: " eu ",
    countryCode: " fi ",
    hostname: " node-fi-1 ",
    expiresInMinutes: "45",
  });

  assert(input.name === "finland-1", "expected trimmed name");
  assert(input.region === "eu", "expected trimmed region");
  assert(input.country_code === "FI", "expected uppercase country code");
  assert(input.hostname === "node-fi-1", "expected trimmed hostname");
  assert(input.expires_in_minutes === 45, "expected parsed expiry");
}

function formatsTimestampsSafely(): void {
  assert(formatNodeTimestamp(null) === "-", "expected null timestamp fallback");
  assert(formatNodeTimestamp("not-a-date") === "-", "expected invalid timestamp fallback");
  assert(formatNodeTimestamp("2026-05-16T02:46:28Z") !== "-", "expected formatted timestamp");
}

function formatsRuntimeEventTypes(): void {
  assert(formatRuntimeEventType(null) === "Runtime event", "expected empty runtime event fallback");
  assert(formatRuntimeEventType("dry_run_failure") === "Dry Run Failure", "expected readable dry-run event type");
  assert(formatRuntimeEventType("process_prepare_start_intent") === "Process Prepare Start Intent", "expected readable process intent event type");
}

function checksLifecycleActions(): void {
  assert(canDrain({ status: "active", drain_state: "active" }), "expected active node to be drainable");
  assert(!canDrain({ status: "disabled", drain_state: "active" }), "expected disabled node not drainable");
  assert(!canDrain({ status: "active", drain_state: "draining" }), "expected draining node not drainable");
  assert(canUndrain({ status: "active", drain_state: "draining" }), "expected draining node undrainable");
  assert(!canUndrain({ status: "active", drain_state: "active" }), "expected active drain state not undrainable");
  assert(canDisable({ status: "active" }), "expected active node disableable");
  assert(!canDisable({ status: "disabled" }), "expected disabled node not disableable");
  assert(canEnable({ status: "disabled" }), "expected disabled node enableable");
  assert(!canEnable({ status: "active" }), "expected active node not enableable");
}

function formatsConfigRevisionDisplay(): void {
  assert(configRevisionStatusClass("applied") === "status-active", "expected applied status class");
  assert(configRevisionStatusClass("failed") === "status-disabled", "expected failed status class");
  assert(configRevisionStatusClass("rolled_back") === "status-archived", "expected rolled back status class");
  assert(configRevisionStatusClass("pending") === "status-pending", "expected pending status class");
  assert(formatConfigRevisionBundle(null) === "-", "expected empty bundle fallback");
  assert(formatConfigRevisionBundle("raw") === "raw", "expected string bundle passthrough");
  assert(formatConfigRevisionBundle({ revision_number: 2 }).includes('"revision_number": 2'), "expected JSON bundle formatting");
}

runTests();
