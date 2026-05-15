import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import {
  activateUser,
  createUser,
  listUsers,
  PanelApiError,
  suspendUser,
  updateUser,
  type User,
} from "../lib/api";
import type { StoredSession } from "../lib/session";
import {
  buildCreateUserInput,
  buildUpdateUserInput,
  emptyUserForm,
  userToForm,
  validateUserForm,
  type UserFormState,
} from "../lib/userForm";

interface UsersPageProps {
  session: StoredSession;
  onUnauthorized: () => void;
}

type LoadState = "idle" | "loading" | "loaded" | "failed";
type FormMode = "create" | "edit";

export function UsersPage({ session, onUnauthorized }: UsersPageProps) {
  const [users, setUsers] = useState<User[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("idle");
  const [formMode, setFormMode] = useState<FormMode>("create");
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [formState, setFormState] = useState<UserFormState>(() => emptyUserForm());
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [isMutating, setIsMutating] = useState(false);
  const [mutatingUserID, setMutatingUserID] = useState<string | null>(null);

  const activeUsers = useMemo(() => users.filter((user) => user.status === "active").length, [users]);

  const loadUsers = useCallback(async () => {
    setLoadState("loading");
    setErrorMessage(null);

    try {
      const loadedUsers = await listUsers(session);
      setUsers(loadedUsers);
      setLoadState("loaded");
    } catch (error) {
      if (handleUnauthorizedError(error, onUnauthorized)) {
        return;
      }
      setErrorMessage(formatPanelError(error, "Unable to load users."));
      setLoadState("failed");
    }
  }, [onUnauthorized, session]);

  useEffect(() => {
    let isMounted = true;

    async function loadInitialUsers() {
      setLoadState("loading");
      setErrorMessage(null);

      try {
        const loadedUsers = await listUsers(session);

        if (!isMounted) {
          return;
        }

        setUsers(loadedUsers);
        setLoadState("loaded");
      } catch (error) {
        if (!isMounted) {
          return;
        }

        if (handleUnauthorizedError(error, onUnauthorized)) {
          return;
        }

        setErrorMessage(formatPanelError(error, "Unable to load users."));
        setLoadState("failed");
      }
    }

    loadInitialUsers();

    return () => {
      isMounted = false;
    };
  }, [onUnauthorized, session]);

  function updateFormField(fieldName: keyof UserFormState, value: string) {
    setFormState((currentValue) => ({ ...currentValue, [fieldName]: value }));
  }

  function resetForm(message?: string) {
    setFormMode("create");
    setEditingUser(null);
    setFormState(emptyUserForm());
    setSuccessMessage(message ?? null);
  }

  function startEdit(user: User) {
    setFormMode("edit");
    setEditingUser(user);
    setFormState(userToForm(user));
    setErrorMessage(null);
    setSuccessMessage(null);
  }

  async function submitUserForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const validationError = validateUserForm(formState);
    if (validationError) {
      setErrorMessage(validationError);
      setSuccessMessage(null);
      return;
    }

    setIsMutating(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      if (formMode === "edit" && editingUser) {
        await updateUser(session, editingUser.id, buildUpdateUserInput(formState));
        resetForm("User updated.");
      } else {
        await createUser(session, buildCreateUserInput(formState));
        resetForm("User created.");
      }
      await loadUsers();
    } catch (error) {
      if (handleUnauthorizedError(error, onUnauthorized)) {
        return;
      }
      setErrorMessage(formatPanelError(error, "Unable to save user."));
    } finally {
      setIsMutating(false);
    }
  }

  async function updateUserStatus(user: User, action: "suspend" | "activate") {
    setMutatingUserID(user.id);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      if (action === "suspend") {
        await suspendUser(session, user.id);
        setSuccessMessage("User suspended.");
      } else {
        await activateUser(session, user.id);
        setSuccessMessage("User activated.");
      }
      await loadUsers();
    } catch (error) {
      if (handleUnauthorizedError(error, onUnauthorized)) {
        return;
      }
      setErrorMessage(formatPanelError(error, "Unable to update user status."));
    } finally {
      setMutatingUserID(null);
    }
  }

  return (
    <div className="page-stack" id="users">
      <section className="page-header">
        <div>
          <p className="eyebrow">Users</p>
          <h2>Users</h2>
          <p>Create, edit, suspend, and activate provider users through the panel-api admin API.</p>
        </div>
        <div className="header-actions">
          <span className="pill">{users.length} total</span>
          <span className="pill">{activeUsers} active</span>
        </div>
      </section>

      <section className="management-grid">
        <form className="user-form-panel" onSubmit={submitUserForm}>
          <div className="section-heading">
            <div>
              <p className="eyebrow">{formMode === "edit" ? "Edit user" : "New user"}</p>
              <h3>{formMode === "edit" ? editingUser?.email : "Create user"}</h3>
            </div>
            {formMode === "edit" ? (
              <button className="ghost-button" type="button" onClick={() => resetForm()} disabled={isMutating}>
                Cancel
              </button>
            ) : null}
          </div>

          <label className="field-label" htmlFor="user-email">
            Email
          </label>
          <input
            id="user-email"
            className="text-field"
            type="email"
            autoComplete="off"
            value={formState.email}
            onChange={(event) => updateFormField("email", event.target.value)}
          />

          <label className="field-label" htmlFor="user-display-name">
            Display name
          </label>
          <input
            id="user-display-name"
            className="text-field"
            type="text"
            autoComplete="off"
            value={formState.displayName}
            onChange={(event) => updateFormField("displayName", event.target.value)}
          />

          <button className="primary-button" type="submit" disabled={isMutating}>
            {isMutating ? "Saving..." : formMode === "edit" ? "Save changes" : "Create user"}
          </button>
        </form>

        <div className="users-feedback-panel">
          <p className="eyebrow">State</p>
          {loadState === "loading" ? <p className="state-text">Loading users...</p> : null}
          {loadState === "failed" ? <p className="error-text">{errorMessage}</p> : null}
          {loadState === "loaded" && !errorMessage && !successMessage ? (
            <p className="state-text">Users list is ready.</p>
          ) : null}
          {errorMessage && loadState !== "failed" ? <p className="error-text">{errorMessage}</p> : null}
          {successMessage ? <p className="success-text">{successMessage}</p> : null}
          <button className="secondary-button" type="button" onClick={loadUsers} disabled={loadState === "loading"}>
            Refresh
          </button>
        </div>
      </section>

      {loadState === "loaded" && users.length === 0 ? <p className="state-card">No users yet. Create the first user above.</p> : null}

      {users.length > 0 ? (
        <div className="table-wrap">
          <table className="data-table users-table">
            <thead>
              <tr>
                <th>Email</th>
                <th>Display name</th>
                <th>Status</th>
                <th>ID</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id}>
                  <td>{user.email}</td>
                  <td>{user.display_name || "-"}</td>
                  <td>
                    <span className={`status-badge status-${user.status}`}>{user.status}</span>
                  </td>
                  <td className="mono-cell">{user.id}</td>
                  <td>
                    <div className="row-actions">
                      <button className="table-button" type="button" onClick={() => startEdit(user)} disabled={isMutating}>
                        Edit
                      </button>
                      {user.status === "active" ? (
                        <button
                          className="table-button danger"
                          type="button"
                          onClick={() => updateUserStatus(user, "suspend")}
                          disabled={mutatingUserID === user.id}
                        >
                          {mutatingUserID === user.id ? "Suspending..." : "Suspend"}
                        </button>
                      ) : (
                        <button
                          className="table-button"
                          type="button"
                          onClick={() => updateUserStatus(user, "activate")}
                          disabled={mutatingUserID === user.id}
                        >
                          {mutatingUserID === user.id ? "Activating..." : "Activate"}
                        </button>
                      )}
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
