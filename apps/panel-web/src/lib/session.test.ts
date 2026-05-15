import { clearStoredSession, loadStoredSession, saveStoredSession, SESSION_STORAGE_KEY, type StoredSession } from "./session";

class MemoryStorage {
  private values = new Map<string, string>();

  getItem(key: string): string | null {
    return this.values.get(key) ?? null;
  }

  setItem(key: string, value: string): void {
    this.values.set(key, value);
  }

  removeItem(key: string): void {
    this.values.delete(key);
  }
}

const now = new Date("2026-05-15T10:00:00.000Z");

function testSession(expiresAt = "2026-05-15T11:00:00.000Z"): StoredSession {
  return {
    admin: {
      id: "admin-1",
      email: "owner@example.com",
      status: "active",
      two_factor_enabled: false,
      created_at: "2026-05-15T09:00:00.000Z",
      updated_at: "2026-05-15T09:00:00.000Z",
      last_login_at: null,
    },
    session: {
      id: "session-1",
      admin_id: "admin-1",
      token: "session-token",
      expires_at: expiresAt,
      created_at: "2026-05-15T09:00:00.000Z",
    },
  };
}

function assert(condition: boolean, message: string): void {
  if (!condition) {
    throw new Error(message);
  }
}

function runTests(): void {
  loadsValidSessionFromSessionStorage();
  rejectsExpiredSession();
  rejectsMalformedSession();
  clearsLegacyLocalStorage();
  clearsStoredSessionFromBothStores();
}

function loadsValidSessionFromSessionStorage(): void {
  const sessionStorage = new MemoryStorage();
  const localStorage = new MemoryStorage();
  const expected = testSession();

  saveStoredSession(expected, sessionStorage, localStorage);
  const loaded = loadStoredSession(sessionStorage, localStorage, now);

  assert(loaded?.session.token === "session-token", "expected valid session to load from sessionStorage");
  assert(sessionStorage.getItem(SESSION_STORAGE_KEY) !== null, "expected sessionStorage value to remain");
}

function rejectsExpiredSession(): void {
  const sessionStorage = new MemoryStorage();
  const localStorage = new MemoryStorage();
  sessionStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(testSession("2026-05-15T09:59:59.000Z")));

  const loaded = loadStoredSession(sessionStorage, localStorage, now);

  assert(loaded === null, "expected expired session to be rejected");
  assert(sessionStorage.getItem(SESSION_STORAGE_KEY) === null, "expected expired session to be cleared");
}

function rejectsMalformedSession(): void {
  const sessionStorage = new MemoryStorage();
  const localStorage = new MemoryStorage();
  sessionStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify({ admin: { email: "owner@example.com" } }));

  const loaded = loadStoredSession(sessionStorage, localStorage, now);

  assert(loaded === null, "expected malformed session to be rejected");
  assert(sessionStorage.getItem(SESSION_STORAGE_KEY) === null, "expected malformed session to be cleared");
}

function clearsLegacyLocalStorage(): void {
  const sessionStorage = new MemoryStorage();
  const localStorage = new MemoryStorage();
  localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(testSession()));

  const loaded = loadStoredSession(sessionStorage, localStorage, now);

  assert(loaded === null, "expected legacy localStorage value not to restore a session");
  assert(localStorage.getItem(SESSION_STORAGE_KEY) === null, "expected legacy localStorage key to be cleared");
}

function clearsStoredSessionFromBothStores(): void {
  const sessionStorage = new MemoryStorage();
  const localStorage = new MemoryStorage();
  sessionStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(testSession()));
  localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(testSession()));

  clearStoredSession(sessionStorage, localStorage);

  assert(sessionStorage.getItem(SESSION_STORAGE_KEY) === null, "expected sessionStorage to be cleared");
  assert(localStorage.getItem(SESSION_STORAGE_KEY) === null, "expected legacy localStorage to be cleared");
}

runTests();
