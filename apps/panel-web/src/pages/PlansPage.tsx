import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { archivePlan, createPlan, listPlans, PanelApiError, updatePlan, type Plan } from "../lib/api";
import {
  buildCreatePlanInput,
  buildUpdatePlanInput,
  emptyPlanForm,
  planToForm,
  validatePlanForm,
  type PlanFormState,
} from "../lib/planForm";
import type { StoredSession } from "../lib/session";

interface PlansPageProps {
  session: StoredSession;
  onUnauthorized: () => void;
}

type LoadState = "idle" | "loading" | "loaded" | "failed";
type FormMode = "create" | "edit";

export function PlansPage({ session, onUnauthorized }: PlansPageProps) {
  const [plans, setPlans] = useState<Plan[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("idle");
  const [formMode, setFormMode] = useState<FormMode>("create");
  const [editingPlan, setEditingPlan] = useState<Plan | null>(null);
  const [formState, setFormState] = useState<PlanFormState>(() => emptyPlanForm());
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [isMutating, setIsMutating] = useState(false);
  const [mutatingPlanID, setMutatingPlanID] = useState<string | null>(null);

  const activePlans = useMemo(() => plans.filter((plan) => plan.status === "active").length, [plans]);

  const loadPlans = useCallback(async () => {
    setLoadState("loading");
    setErrorMessage(null);

    try {
      const loadedPlans = await listPlans(session);
      setPlans(loadedPlans);
      setLoadState("loaded");
    } catch (error) {
      if (handleUnauthorizedError(error, onUnauthorized)) {
        return;
      }
      setErrorMessage(formatPanelError(error, "Unable to load plans."));
      setLoadState("failed");
    }
  }, [onUnauthorized, session]);

  useEffect(() => {
    let isMounted = true;

    async function loadInitialPlans() {
      setLoadState("loading");
      setErrorMessage(null);

      try {
        const loadedPlans = await listPlans(session);

        if (!isMounted) {
          return;
        }

        setPlans(loadedPlans);
        setLoadState("loaded");
      } catch (error) {
        if (!isMounted) {
          return;
        }

        if (handleUnauthorizedError(error, onUnauthorized)) {
          return;
        }

        setErrorMessage(formatPanelError(error, "Unable to load plans."));
        setLoadState("failed");
      }
    }

    loadInitialPlans();

    return () => {
      isMounted = false;
    };
  }, [onUnauthorized, session]);

  function updateFormField(fieldName: keyof PlanFormState, value: string | boolean) {
    setFormState((currentValue) => ({ ...currentValue, [fieldName]: value }));
  }

  function resetForm(message?: string) {
    setFormMode("create");
    setEditingPlan(null);
    setFormState(emptyPlanForm());
    setSuccessMessage(message ?? null);
  }

  function startEdit(plan: Plan) {
    setFormMode("edit");
    setEditingPlan(plan);
    setFormState(planToForm(plan));
    setErrorMessage(null);
    setSuccessMessage(null);
  }

  async function submitPlanForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const validationError = validatePlanForm(formState);
    if (validationError) {
      setErrorMessage(validationError);
      setSuccessMessage(null);
      return;
    }

    setIsMutating(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      if (formMode === "edit" && editingPlan) {
        await updatePlan(session, editingPlan.id, buildUpdatePlanInput(formState));
        resetForm("Plan updated.");
      } else {
        await createPlan(session, buildCreatePlanInput(formState));
        resetForm("Plan created.");
      }
      await loadPlans();
    } catch (error) {
      if (handleUnauthorizedError(error, onUnauthorized)) {
        return;
      }
      setErrorMessage(formatPanelError(error, "Unable to save plan."));
    } finally {
      setIsMutating(false);
    }
  }

  async function archiveSelectedPlan(plan: Plan) {
    setMutatingPlanID(plan.id);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      await archivePlan(session, plan.id);
      setSuccessMessage("Plan archived.");
      if (editingPlan?.id === plan.id) {
        resetForm("Plan archived.");
      }
      await loadPlans();
    } catch (error) {
      if (handleUnauthorizedError(error, onUnauthorized)) {
        return;
      }
      setErrorMessage(formatPanelError(error, "Unable to archive plan."));
    } finally {
      setMutatingPlanID(null);
    }
  }

  return (
    <div className="page-stack" id="plans">
      <section className="page-header">
        <div>
          <p className="eyebrow">Plans</p>
          <h2>Plans</h2>
          <p>Create, edit, and archive subscription plans through the panel-api admin API.</p>
        </div>
        <div className="header-actions">
          <span className="pill">{plans.length} total</span>
          <span className="pill">{activePlans} active</span>
        </div>
      </section>

      <section className="management-grid">
        <form className="management-panel" onSubmit={submitPlanForm}>
          <div className="section-heading">
            <div>
              <p className="eyebrow">{formMode === "edit" ? "Edit plan" : "New plan"}</p>
              <h3>{formMode === "edit" ? editingPlan?.name : "Create plan"}</h3>
            </div>
            {formMode === "edit" ? (
              <button className="ghost-button" type="button" onClick={() => resetForm()} disabled={isMutating}>
                Cancel
              </button>
            ) : null}
          </div>

          <label className="field-label" htmlFor="plan-name">
            Name
          </label>
          <input
            id="plan-name"
            className="text-field"
            type="text"
            autoComplete="off"
            value={formState.name}
            onChange={(event) => updateFormField("name", event.target.value)}
          />

          <div className="form-grid">
            <div>
              <label className="field-label" htmlFor="plan-duration-days">
                Duration days
              </label>
              <input
                id="plan-duration-days"
                className="text-field"
                type="number"
                min="1"
                inputMode="numeric"
                value={formState.durationDays}
                onChange={(event) => updateFormField("durationDays", event.target.value)}
              />
            </div>
            <div>
              <label className="field-label" htmlFor="plan-device-limit">
                Device limit
              </label>
              <input
                id="plan-device-limit"
                className="text-field"
                type="number"
                min="1"
                inputMode="numeric"
                value={formState.deviceLimit}
                onChange={(event) => updateFormField("deviceLimit", event.target.value)}
              />
            </div>
          </div>

          <label className="check-row" htmlFor="plan-has-traffic-limit">
            <input
              id="plan-has-traffic-limit"
              type="checkbox"
              checked={formState.hasTrafficLimit}
              onChange={(event) => updateFormField("hasTrafficLimit", event.target.checked)}
            />
            <span>Set traffic limit</span>
          </label>

          {formState.hasTrafficLimit ? (
            <>
              <label className="field-label" htmlFor="plan-traffic-limit">
                Traffic limit bytes
              </label>
              <input
                id="plan-traffic-limit"
                className="text-field"
                type="number"
                min="1"
                inputMode="numeric"
                value={formState.trafficLimitBytes}
                onChange={(event) => updateFormField("trafficLimitBytes", event.target.value)}
              />
            </>
          ) : null}

          <button className="primary-button" type="submit" disabled={isMutating}>
            {isMutating ? "Saving..." : formMode === "edit" ? "Save changes" : "Create plan"}
          </button>
        </form>

        <div className="feedback-panel">
          <p className="eyebrow">State</p>
          {loadState === "loading" ? <p className="state-text">Loading plans...</p> : null}
          {loadState === "failed" ? <p className="error-text">{errorMessage}</p> : null}
          {loadState === "loaded" && !errorMessage && !successMessage ? (
            <p className="state-text">Plans list is ready.</p>
          ) : null}
          {errorMessage && loadState !== "failed" ? <p className="error-text">{errorMessage}</p> : null}
          {successMessage ? <p className="success-text">{successMessage}</p> : null}
          <button className="secondary-button" type="button" onClick={loadPlans} disabled={loadState === "loading"}>
            Refresh
          </button>
        </div>
      </section>

      {loadState === "loaded" && plans.length === 0 ? <p className="state-card">No plans yet. Create the first plan above.</p> : null}

      {plans.length > 0 ? (
        <div className="table-wrap">
          <table className="data-table management-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Duration</th>
                <th>Devices</th>
                <th>Traffic</th>
                <th>Status</th>
                <th>ID</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {plans.map((plan) => (
                <tr key={plan.id}>
                  <td>{plan.name}</td>
                  <td>{plan.duration_days} days</td>
                  <td>{plan.device_limit}</td>
                  <td>{formatTrafficLimit(plan.traffic_limit_bytes)}</td>
                  <td>
                    <span className={`status-badge status-${plan.status}`}>{plan.status}</span>
                  </td>
                  <td className="mono-cell">{plan.id}</td>
                  <td>
                    <div className="row-actions">
                      <button className="table-button" type="button" onClick={() => startEdit(plan)} disabled={isMutating}>
                        Edit
                      </button>
                      <button
                        className="table-button danger"
                        type="button"
                        onClick={() => archiveSelectedPlan(plan)}
                        disabled={plan.status === "archived" || mutatingPlanID === plan.id}
                      >
                        {mutatingPlanID === plan.id ? "Archiving..." : plan.status === "archived" ? "Archived" : "Archive"}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </div>
  );
}

function formatTrafficLimit(value: number | null): string {
  if (value === null) {
    return "Unlimited";
  }
  return new Intl.NumberFormat(undefined).format(value);
}

function handleUnauthorizedError(error: unknown, onUnauthorized: () => void): boolean {
  if (error instanceof PanelApiError && error.status === 401) {
    onUnauthorized();
    return true;
  }
  return false;
}

function formatPanelError(error: unknown, fallbackMessage: string): string {
  if (error instanceof PanelApiError) {
    return `${error.message} (${error.code})`;
  }
  return error instanceof Error ? error.message : fallbackMessage;
}
