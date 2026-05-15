import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import {
  createSubscription,
  listPlans,
  listSubscriptions,
  listUsers,
  PanelApiError,
  renewSubscription,
  updateSubscription,
  type Plan,
  type Subscription,
  type User,
} from "../lib/api";
import type { StoredSession } from "../lib/session";
import {
  buildCreateSubscriptionInput,
  buildRenewSubscriptionInput,
  buildUpdateSubscriptionInput,
  emptySubscriptionForm,
  subscriptionToForm,
  validateCreateSubscriptionForm,
  validateRenewSubscriptionForm,
  validateUpdateSubscriptionForm,
  type SubscriptionFormState,
  type SubscriptionStatus,
} from "../lib/subscriptionForm";

interface SubscriptionsPageProps {
  session: StoredSession;
  onUnauthorized: () => void;
}

type LoadState = "idle" | "loading" | "loaded" | "failed";
type FormMode = "create" | "edit";

export function SubscriptionsPage({ session, onUnauthorized }: SubscriptionsPageProps) {
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [plans, setPlans] = useState<Plan[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("idle");
  const [formMode, setFormMode] = useState<FormMode>("create");
  const [editingSubscription, setEditingSubscription] = useState<Subscription | null>(null);
  const [formState, setFormState] = useState<SubscriptionFormState>(() => emptySubscriptionForm());
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [isMutating, setIsMutating] = useState(false);
  const [mutatingSubscriptionID, setMutatingSubscriptionID] = useState<string | null>(null);

  const activeSubscriptions = useMemo(
    () => subscriptions.filter((subscription) => subscription.status === "active").length,
    [subscriptions],
  );
  const activePlans = useMemo(() => plans.filter((plan) => plan.status === "active"), [plans]);

  const loadPageData = useCallback(async () => {
    setLoadState("loading");
    setErrorMessage(null);

    try {
      const [loadedSubscriptions, loadedUsers, loadedPlans] = await Promise.all([
        listSubscriptions(session),
        listUsers(session),
        listPlans(session),
      ]);
      setSubscriptions(loadedSubscriptions);
      setUsers(loadedUsers);
      setPlans(loadedPlans);
      setLoadState("loaded");
    } catch (error) {
      if (handleUnauthorizedError(error, onUnauthorized)) {
        return;
      }
      setErrorMessage(formatPanelError(error, "Unable to load subscriptions."));
      setLoadState("failed");
    }
  }, [onUnauthorized, session]);

  useEffect(() => {
    let isMounted = true;

    async function loadInitialPageData() {
      setLoadState("loading");
      setErrorMessage(null);

      try {
        const [loadedSubscriptions, loadedUsers, loadedPlans] = await Promise.all([
          listSubscriptions(session),
          listUsers(session),
          listPlans(session),
        ]);

        if (!isMounted) {
          return;
        }

        setSubscriptions(loadedSubscriptions);
        setUsers(loadedUsers);
        setPlans(loadedPlans);
        setLoadState("loaded");
      } catch (error) {
        if (!isMounted) {
          return;
        }

        if (handleUnauthorizedError(error, onUnauthorized)) {
          return;
        }

        setErrorMessage(formatPanelError(error, "Unable to load subscriptions."));
        setLoadState("failed");
      }
    }

    loadInitialPageData();

    return () => {
      isMounted = false;
    };
  }, [onUnauthorized, session]);

  function updateFormField(fieldName: keyof SubscriptionFormState, value: string | boolean) {
    setFormState((currentValue) => ({ ...currentValue, [fieldName]: value }));
  }

  function resetForm(message?: string) {
    setFormMode("create");
    setEditingSubscription(null);
    setFormState(emptySubscriptionForm());
    setSuccessMessage(message ?? null);
  }

  function startEdit(subscription: Subscription) {
    setFormMode("edit");
    setEditingSubscription(subscription);
    setFormState(subscriptionToForm(subscription));
    setErrorMessage(null);
    setSuccessMessage(null);
  }

  async function submitSubscriptionForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const validationError =
      formMode === "edit" ? validateUpdateSubscriptionForm(formState) : validateCreateSubscriptionForm(formState);
    if (validationError) {
      setErrorMessage(validationError);
      setSuccessMessage(null);
      return;
    }

    setIsMutating(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      if (formMode === "edit" && editingSubscription) {
        await updateSubscription(session, editingSubscription.id, buildUpdateSubscriptionInput(formState));
        resetForm("Subscription updated.");
      } else {
        await createSubscription(session, buildCreateSubscriptionInput(formState));
        resetForm("Subscription created.");
      }
      await loadPageData();
    } catch (error) {
      if (handleUnauthorizedError(error, onUnauthorized)) {
        return;
      }
      setErrorMessage(formatPanelError(error, "Unable to save subscription."));
    } finally {
      setIsMutating(false);
    }
  }

  async function renewSelectedSubscription(subscription: Subscription) {
    const validationError = validateRenewSubscriptionForm(formState);
    if (validationError) {
      setErrorMessage(validationError);
      setSuccessMessage(null);
      return;
    }

    setMutatingSubscriptionID(subscription.id);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      await renewSubscription(session, subscription.id, buildRenewSubscriptionInput(formState));
      setSuccessMessage("Subscription renewed.");
      await loadPageData();
    } catch (error) {
      if (handleUnauthorizedError(error, onUnauthorized)) {
        return;
      }
      setErrorMessage(formatPanelError(error, "Unable to renew subscription."));
    } finally {
      setMutatingSubscriptionID(null);
    }
  }

  return (
    <div className="page-stack" id="subscriptions">
      <section className="page-header">
        <div>
          <p className="eyebrow">Subscriptions</p>
          <h2>Subscriptions</h2>
          <p>Create, update, and renew subscriptions through the panel-api admin API.</p>
        </div>
        <div className="header-actions">
          <span className="pill">{subscriptions.length} total</span>
          <span className="pill">{activeSubscriptions} active</span>
        </div>
      </section>

      <section className="management-grid">
        <form className="management-panel" onSubmit={submitSubscriptionForm}>
          <div className="section-heading">
            <div>
              <p className="eyebrow">{formMode === "edit" ? "Edit subscription" : "New subscription"}</p>
              <h3>{formMode === "edit" ? editingSubscription?.id : "Create subscription"}</h3>
            </div>
            {formMode === "edit" ? (
              <button className="ghost-button" type="button" onClick={() => resetForm()} disabled={isMutating}>
                Cancel
              </button>
            ) : null}
          </div>

          {formMode === "create" ? (
            <>
              <label className="field-label" htmlFor="subscription-user">
                User
              </label>
              <select
                id="subscription-user"
                className="select-field"
                value={formState.userID}
                onChange={(event) => updateFormField("userID", event.target.value)}
              >
                <option value="">Select user</option>
                {users.map((user) => (
                  <option key={user.id} value={user.id}>
                    {user.email}
                  </option>
                ))}
              </select>

              <label className="field-label" htmlFor="subscription-plan">
                Plan
              </label>
              <select
                id="subscription-plan"
                className="select-field"
                value={formState.planID}
                onChange={(event) => updateFormField("planID", event.target.value)}
              >
                <option value="">Select plan</option>
                {activePlans.map((plan) => (
                  <option key={plan.id} value={plan.id}>
                    {plan.name}
                  </option>
                ))}
              </select>
            </>
          ) : (
            <>
              <label className="field-label" htmlFor="subscription-status">
                Status
              </label>
              <select
                id="subscription-status"
                className="select-field"
                value={formState.status}
                onChange={(event) => updateFormField("status", event.target.value as SubscriptionStatus)}
              >
                <option value="active">active</option>
                <option value="expired">expired</option>
                <option value="suspended">suspended</option>
              </select>

              <label className="field-label" htmlFor="subscription-device-limit">
                Device limit
              </label>
              <input
                id="subscription-device-limit"
                className="text-field"
                type="number"
                min="1"
                inputMode="numeric"
                value={formState.deviceLimit}
                onChange={(event) => updateFormField("deviceLimit", event.target.value)}
              />

              <label className="check-row" htmlFor="subscription-has-traffic-limit">
                <input
                  id="subscription-has-traffic-limit"
                  type="checkbox"
                  checked={formState.hasTrafficLimit}
                  onChange={(event) => updateFormField("hasTrafficLimit", event.target.checked)}
                />
                <span>Set traffic limit</span>
              </label>

              {formState.hasTrafficLimit ? (
                <>
                  <label className="field-label" htmlFor="subscription-traffic-limit">
                    Traffic limit bytes
                  </label>
                  <input
                    id="subscription-traffic-limit"
                    className="text-field"
                    type="number"
                    min="1"
                    inputMode="numeric"
                    value={formState.trafficLimitBytes}
                    onChange={(event) => updateFormField("trafficLimitBytes", event.target.value)}
                  />
                </>
              ) : null}
            </>
          )}

          <label className="check-row" htmlFor="subscription-has-preferred-region">
            <input
              id="subscription-has-preferred-region"
              type="checkbox"
              checked={formState.hasPreferredRegion}
              onChange={(event) => updateFormField("hasPreferredRegion", event.target.checked)}
            />
            <span>Set preferred region</span>
          </label>

          {formState.hasPreferredRegion ? (
            <>
              <label className="field-label" htmlFor="subscription-preferred-region">
                Preferred region
              </label>
              <input
                id="subscription-preferred-region"
                className="text-field"
                type="text"
                autoComplete="off"
                value={formState.preferredRegion}
                onChange={(event) => updateFormField("preferredRegion", event.target.value)}
              />
            </>
          ) : null}

          <label className="field-label" htmlFor="subscription-renew-days">
            Renew days
          </label>
          <input
            id="subscription-renew-days"
            className="text-field"
            type="number"
            min="1"
            inputMode="numeric"
            value={formState.renewDays}
            onChange={(event) => updateFormField("renewDays", event.target.value)}
          />

          <button className="primary-button" type="submit" disabled={isMutating}>
            {isMutating ? "Saving..." : formMode === "edit" ? "Save changes" : "Create subscription"}
          </button>
        </form>

        <div className="feedback-panel">
          <p className="eyebrow">State</p>
          {loadState === "loading" ? <p className="state-text">Loading subscriptions...</p> : null}
          {loadState === "failed" ? <p className="error-text">{errorMessage}</p> : null}
          {loadState === "loaded" && !errorMessage && !successMessage ? (
            <p className="state-text">Subscriptions list is ready.</p>
          ) : null}
          {errorMessage && loadState !== "failed" ? <p className="error-text">{errorMessage}</p> : null}
          {successMessage ? <p className="success-text">{successMessage}</p> : null}
          <button className="secondary-button" type="button" onClick={loadPageData} disabled={loadState === "loading"}>
            Refresh
          </button>
        </div>
      </section>

      {loadState === "loaded" && subscriptions.length === 0 ? (
        <p className="state-card">No subscriptions yet. Create the first subscription above.</p>
      ) : null}

      {subscriptions.length > 0 ? (
        <div className="table-wrap">
          <table className="data-table subscriptions-table">
            <thead>
              <tr>
                <th>User</th>
                <th>Plan</th>
                <th>Status</th>
                <th>Expires</th>
                <th>Traffic</th>
                <th>Devices</th>
                <th>Region</th>
                <th>ID</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {subscriptions.map((subscription) => (
                <tr key={subscription.id}>
                  <td>{userLabel(users, subscription.user_id)}</td>
                  <td>{planLabel(plans, subscription.plan_id)}</td>
                  <td>
                    <span className={`status-badge status-${subscription.status}`}>{subscription.status}</span>
                  </td>
                  <td>{formatDate(subscription.expires_at)}</td>
                  <td>{formatTraffic(subscription.traffic_used_bytes, subscription.traffic_limit_bytes)}</td>
                  <td>{subscription.device_limit}</td>
                  <td>{subscription.preferred_region || "-"}</td>
                  <td className="mono-cell">{subscription.id}</td>
                  <td>
                    <div className="row-actions">
                      <button className="table-button" type="button" onClick={() => startEdit(subscription)} disabled={isMutating}>
                        Edit
                      </button>
                      <button
                        className="table-button"
                        type="button"
                        onClick={() => renewSelectedSubscription(subscription)}
                        disabled={mutatingSubscriptionID === subscription.id}
                      >
                        {mutatingSubscriptionID === subscription.id ? "Renewing..." : `Renew ${formState.renewDays || "?"}d`}
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

function userLabel(users: User[], userID: string): string {
  return users.find((user) => user.id === userID)?.email ?? userID;
}

function planLabel(plans: Plan[], planID: string): string {
  return plans.find((plan) => plan.id === planID)?.name ?? planID;
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function formatTraffic(usedBytes: number, limitBytes: number | null): string {
  const used = new Intl.NumberFormat(undefined).format(usedBytes);
  if (limitBytes === null) {
    return `${used} / unlimited`;
  }
  return `${used} / ${new Intl.NumberFormat(undefined).format(limitBytes)}`;
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
