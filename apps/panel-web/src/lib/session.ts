const SESSION_STORAGE_KEY = "lenker.panel.session";

export interface AdminUser {
  id: string;
  email: string;
  status: "active" | "suspended";
  two_factor_enabled: boolean;
  created_at: string;
  updated_at: string;
  last_login_at?: string | null;
}

export interface AdminSession {
  id: string;
  admin_id: string;
  token: string;
  expires_at: string;
  created_at: string;
}

export interface StoredSession {
  admin: AdminUser;
  session: AdminSession;
}

export function loadStoredSession(): StoredSession | null {
  const rawValue = window.localStorage.getItem(SESSION_STORAGE_KEY);

  if (!rawValue) {
    return null;
  }

  try {
    const parsedValue = JSON.parse(rawValue) as StoredSession;

    if (!parsedValue.admin?.email || !parsedValue.session?.token) {
      clearStoredSession();
      return null;
    }

    return parsedValue;
  } catch {
    clearStoredSession();
    return null;
  }
}

export function saveStoredSession(session: StoredSession): void {
  window.localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session));
}

export function clearStoredSession(): void {
  window.localStorage.removeItem(SESSION_STORAGE_KEY);
}
