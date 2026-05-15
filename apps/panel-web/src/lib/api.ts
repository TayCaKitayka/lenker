import type { StoredSession } from "./session";

const DEFAULT_API_BASE_URL = "http://localhost:8080";

interface LoginResponse {
  data: StoredSession;
}

interface ApiErrorResponse {
  error?: {
    code?: string;
    message?: string;
  };
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
    const errorPayload = payload as ApiErrorResponse | null;

    throw new PanelApiError(
      errorPayload?.error?.message || "Login failed",
      errorPayload?.error?.code || "request_failed",
      response.status,
    );
  }

  const loginPayload = payload as LoginResponse | null;

  if (!loginPayload?.data?.admin || !loginPayload.data.session?.token) {
    throw new PanelApiError("Unexpected login response", "invalid_response", response.status);
  }

  return loginPayload.data;
}
