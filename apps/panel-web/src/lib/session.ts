export const SESSION_STORAGE_KEY = "lenker.panel.session";

interface StorageLike {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
  removeItem(key: string): void;
}

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

export function loadStoredSession(
  storage: StorageLike = window.sessionStorage,
  legacyStorage: StorageLike = window.localStorage,
  now: Date = new Date(),
): StoredSession | null {
  const rawValue = storage.getItem(SESSION_STORAGE_KEY);

  legacyStorage.removeItem(SESSION_STORAGE_KEY);

  if (!rawValue) {
    return null;
  }

  try {
    const parsedValue = JSON.parse(rawValue) as unknown;

    if (!isStoredSession(parsedValue, now)) {
      clearStoredSession(storage, legacyStorage);
      return null;
    }

    return parsedValue;
  } catch {
    clearStoredSession(storage, legacyStorage);
    return null;
  }
}

export function saveStoredSession(
  session: StoredSession,
  storage: StorageLike = window.sessionStorage,
  legacyStorage: StorageLike = window.localStorage,
): void {
  storage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session));
  legacyStorage.removeItem(SESSION_STORAGE_KEY);
}

export function clearStoredSession(
  storage: StorageLike = window.sessionStorage,
  legacyStorage: StorageLike = window.localStorage,
): void {
  storage.removeItem(SESSION_STORAGE_KEY);
  legacyStorage.removeItem(SESSION_STORAGE_KEY);
}

function isStoredSession(value: unknown, now: Date): value is StoredSession {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Partial<StoredSession>;
  const admin = candidate.admin;
  const session = candidate.session;

  if (!admin || typeof admin !== "object" || !session || typeof session !== "object") {
    return false;
  }

  if (
    !isNonEmptyString(admin.id) ||
    !isNonEmptyString(admin.email) ||
    !isNonEmptyString(admin.status) ||
    !isNonEmptyString(admin.created_at) ||
    !isNonEmptyString(admin.updated_at)
  ) {
    return false;
  }

  if (
    !isNonEmptyString(session.id) ||
    !isNonEmptyString(session.admin_id) ||
    !isNonEmptyString(session.token) ||
    !isNonEmptyString(session.expires_at) ||
    !isNonEmptyString(session.created_at)
  ) {
    return false;
  }

  const expiresAt = Date.parse(session.expires_at);
  return Number.isFinite(expiresAt) && expiresAt > now.getTime();
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim() !== "";
}
