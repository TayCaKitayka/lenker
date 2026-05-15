import type { StoredSession } from "./session";

const DEFAULT_API_BASE_URL = "http://localhost:8080";

interface LoginResponse {
  data: StoredSession;
}

interface UserListResponse {
  data: User[];
}

interface ApiErrorResponse {
  error?: {
    code?: string;
    message?: string;
  };
}

export interface User {
  id: string;
  email: string;
  status: "active" | "suspended" | "expired";
  display_name: string;
}

export class PanelApiError extends Error {
  readonly code: string;
  readonly status: number;

  constructor(message: string, code: string, status: number) {
    super(message);
    this.name = "PanelApiError";
    this.code = code;
    this.status = status;
  }
}

export function getApiBaseUrl(): string {
  return import.meta.env.VITE_LENKER_PANEL_API_URL || DEFAULT_API_BASE_URL;
}

export async function loginAdmin(email: string, password: string): Promise<StoredSession> {
  const response = await fetch(`${getApiBaseUrl()}/api/v1/auth/admin/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  const payload = (await response.json().catch(() => null)) as LoginResponse | ApiErrorResponse | null;

  if (!response.ok) {
    throwPanelApiError(response, payload, "Login failed");
  }

  const loginPayload = payload as LoginResponse | null;

  if (!loginPayload?.data?.admin || !loginPayload.data.session?.token) {
    throw new PanelApiError("Unexpected login response", "invalid_response", response.status);
  }

  return loginPayload.data;
}

export async function listUsers(session: StoredSession): Promise<User[]> {
  const payload = await authorizedRequest<UserListResponse>(session, "/api/v1/users");
  return payload.data;
}

async function authorizedRequest<TPayload>(session: StoredSession, path: string): Promise<TPayload> {
  const response = await fetch(`${getApiBaseUrl()}${path}`, {
    headers: {
      Authorization: `Bearer ${session.session.token}`,
    },
  });

  const payload = (await response.json().catch(() => null)) as TPayload | ApiErrorResponse | null;

  if (!response.ok) {
    throwPanelApiError(response, payload, "Request failed");
  }

  if (!payload) {
    throw new PanelApiError("Unexpected empty response", "invalid_response", response.status);
  }

  return payload as TPayload;
}

function throwPanelApiError(response: Response, payload: unknown, fallbackMessage: string): never {
  const errorPayload = payload as ApiErrorResponse | null;

  throw new PanelApiError(
    errorPayload?.error?.message || fallbackMessage,
    errorPayload?.error?.code || "request_failed",
    response.status,
  );
}
