import { useEffect, useState } from "react";
import { listUsers, PanelApiError, type User } from "../lib/api";
import type { StoredSession } from "../lib/session";

interface UsersPageProps {
  session: StoredSession;
  onUnauthorized: () => void;
}

type LoadState = "idle" | "loading" | "loaded" | "failed";

export function UsersPage({ session, onUnauthorized }: UsersPageProps) {
  const [users, setUsers] = useState<User[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("idle");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;

    async function loadUsers() {
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

        if (error instanceof PanelApiError && error.status === 401) {
          onUnauthorized();
          return;
        }

        setErrorMessage(error instanceof Error ? error.message : "Unable to load users.");
        setLoadState("failed");
      }
    }

    loadUsers();

    return () => {
      isMounted = false;
    };
  }, [session, onUnauthorized]);

  return (
    <div className="page-stack" id="users">
      <section className="page-header">
        <div>
          <p className="eyebrow">Users</p>
          <h2>Users</h2>
          <p>Read-only user list loaded from panel-api.</p>
        </div>
        <span className="pill">{users.length} total</span>
      </section>

      <section className="surface-card">
        {loadState === "loading" ? <p className="state-text">Loading users...</p> : null}
        {loadState === "failed" ? <p className="error-text">{errorMessage}</p> : null}
        {loadState === "loaded" && users.length === 0 ? <p className="state-text">No users yet.</p> : null}

        {users.length > 0 ? (
          <div className="table-wrap">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Email</th>
                  <th>Display name</th>
                  <th>Status</th>
                  <th>ID</th>
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
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </section>
    </div>
  );
}
