import { useMemo, useState } from "react";
import { clearStoredSession, loadStoredSession, saveStoredSession, type StoredSession } from "../lib/session";

interface LoginFormState {
  email: string;
  secret: string;
}

const initialLoginFormState: LoginFormState = {
  email: "owner@example.com",
  secret: "",
};

export function App() {
  const [storedSession, setStoredSession] = useState<StoredSession | null>(() => loadStoredSession());
  const [formState, setFormState] = useState<LoginFormState>(initialLoginFormState);
  const [message, setMessage] = useState<string | null>(null);

  const expiresAtLabel = useMemo(() => {
    if (!storedSession?.session.expires_at) {
      return "No active session";
    }

    return new Intl.DateTimeFormat(undefined, {
      dateStyle: "medium",
      timeStyle: "short",
    }).format(new Date(storedSession.session.expires_at));
  }, [storedSession]);

  function updateFormField(fieldName: keyof LoginFormState, value: string) {
    setFormState((currentValue) => ({ ...currentValue, [fieldName]: value }));
  }

  function storeManualSession() {
    const trimmedEmail = formState.email.trim();
    const trimmedToken = formState.secret.trim();

    if (!trimmedEmail || !trimmedToken) {
      setMessage("Email and session token are required for the temporary local shell.");
      return;
    }

    const session: StoredSession = {
      admin: {
        id: "local-admin",
        email: trimmedEmail,
        status: "active",
        two_factor_enabled: false,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        last_login_at: new Date().toISOString(),
      },
      session: {
        id: "local-session",
        admin_id: "local-admin",
        token: trimmedToken,
        expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
        created_at: new Date().toISOString(),
      },
    };

    saveStoredSession(session);
    setStoredSession(session);
    setMessage(null);
  }

  function logout() {
    clearStoredSession();
    setStoredSession(null);
    setFormState(initialLoginFormState);
  }

  if (!storedSession) {
    return (
      <main className="auth-layout">
        <section className="auth-card">
          <p className="eyebrow">Lenker Provider Panel</p>
          <h1>Admin access</h1>
          <p className="muted-text">
            First UI foundation for the provider panel. Network login will be wired to panel-api in the next slice.
          </p>

          <label className="field-label" htmlFor="email">
            Admin email
          </label>
          <input
            id="email"
            className="text-field"
            type="email"
            value={formState.email}
            onChange={(event) => updateFormField("email", event.target.value)}
          />

          <label className="field-label" htmlFor="session-token">
            Temporary session token
          </label>
          <input
            id="session-token"
            className="text-field"
            type="password"
            value={formState.secret}
            onChange={(event) => updateFormField("secret", event.target.value)}
          />

          {message ? <p className="error-text">{message}</p> : null}

          <button className="primary-button" type="button" onClick={storeManualSession}>
            Enter dashboard shell
          </button>
        </section>
      </main>
    );
  }

  return (
    <main className="panel-layout">
      <aside className="sidebar">
        <div>
          <p className="eyebrow">Lenker</p>
          <h1>Provider Panel</h1>
        </div>
        <nav className="nav-list" aria-label="Primary navigation">
          <a className="nav-link active" href="#dashboard">Dashboard</a>
          <a className="nav-link" href="#users">Users</a>
          <a className="nav-link" href="#plans">Plans</a>
          <a className="nav-link" href="#subscriptions">Subscriptions</a>
          <a className="nav-link" href="#nodes">Nodes</a>
        </nav>
      </aside>

      <section className="content-shell">
        <header className="topbar">
          <div>
            <p className="muted-text">Signed in as</p>
            <strong>{storedSession.admin.email}</strong>
          </div>
          <button className="secondary-button" type="button" onClick={logout}>
            Sign out
          </button>
        </header>

        <section className="hero-card" id="dashboard">
          <p className="eyebrow">MVP v0.1</p>
          <h2>Dashboard shell is ready</h2>
          <p>
            The React app now has a Vite entrypoint, session persistence, panel layout, and navigation placeholders.
          </p>
          <dl className="details-grid">
            <div>
              <dt>Session expires</dt>
              <dd>{expiresAtLabel}</dd>
            </div>
            <div>
              <dt>Backend target</dt>
              <dd>http://localhost:8080</dd>
            </div>
          </dl>
        </section>

        <section className="cards-grid">
          <StatusCard title="Users" value="Next" description="List, create, suspend, and activate users." />
          <StatusCard title="Plans" value="Next" description="List and maintain subscription plans." />
          <StatusCard title="Subscriptions" value="Next" description="Create, inspect, and renew subscriptions." />
          <StatusCard title="Nodes" value="Next" description="Inspect status, drain, disable, and enable nodes." />
        </section>
      </section>
    </main>
  );
}

interface StatusCardProps {
  title: string;
  value: string;
  description: string;
}

function StatusCard({ title, value, description }: StatusCardProps) {
  return (
    <article className="status-card">
      <p className="muted-text">{title}</p>
      <strong>{value}</strong>
      <span>{description}</span>
    </article>
  );
}
