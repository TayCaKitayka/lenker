import { FormEvent, useMemo, useState } from "react";
import { getApiBaseUrl, loginAdmin, PanelApiError } from "../lib/api";
import { clearStoredSession, loadStoredSession, saveStoredSession, type StoredSession } from "../lib/session";

interface LoginFormState {
  email: string;
  password: string;
}

const initialLoginFormState: LoginFormState = {
  email: "owner@example.com",
  password: "",
};

export function App() {
  const [storedSession, setStoredSession] = useState<StoredSession | null>(() => loadStoredSession());
  const [formState, setFormState] = useState<LoginFormState>(initialLoginFormState);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

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

  async function submitLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const email = formState.email.trim();
    const password = formState.password;

    if (!email || !password) {
      setErrorMessage("Email and password are required.");
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      const session = await loginAdmin(email, password);
      saveStoredSession(session);
      setStoredSession(session);
      setFormState(initialLoginFormState);
    } catch (error) {
      if (error instanceof PanelApiError) {
        setErrorMessage(`${error.message} (${error.code})`);
      } else {
        setErrorMessage("Unable to connect to panel-api.");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  function logout() {
    clearStoredSession();
    setStoredSession(null);
    setFormState(initialLoginFormState);
  }

  if (!storedSession) {
    return (
      <main className="auth-layout">
        <form className="auth-card" onSubmit={submitLogin}>
          <p className="eyebrow">Lenker Provider Panel</p>
          <h1>Admin access</h1>
          <p className="muted-text">
            Sign in with the local admin created by <code>make docker-bootstrap-admin</code>.
          </p>

          <label className="field-label" htmlFor="email">
            Admin email
          </label>
          <input
            id="email"
            className="text-field"
            type="email"
            autoComplete="username"
            value={formState.email}
            onChange={(event) => updateFormField("email", event.target.value)}
          />

          <label className="field-label" htmlFor="password">
            Password
          </label>
          <input
            id="password"
            className="text-field"
            type="password"
            autoComplete="current-password"
            value={formState.password}
            onChange={(event) => updateFormField("password", event.target.value)}
          />

          {errorMessage ? <p className="error-text">{errorMessage}</p> : null}

          <button className="primary-button" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Signing in..." : "Sign in"}
          </button>

          <p className="helper-text">API target: {getApiBaseUrl()}</p>
        </form>
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
            The React app now authenticates against panel-api, stores the admin session locally,
            and renders the first provider dashboard shell.
          </p>
          <dl className="details-grid">
            <div>
              <dt>Session expires</dt>
              <dd>{expiresAtLabel}</dd>
            </div>
            <div>
              <dt>Backend target</dt>
              <dd>{getApiBaseUrl()}</dd>
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
