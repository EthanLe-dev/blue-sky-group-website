import { api, apiResponse, ApiError } from "./api";
export type User = { full_name: string; permissions: string[] };
type Session = { access_token: string; csrf_token: string; user: User };
let session: Session | null = null;
let refreshing: Promise<Session> | null = null;
export async function login(
  email: string,
  password: string,
  turnstile_token: string,
) {
  session = await api<Session>("/admin/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password, turnstile_token }),
  });
  return session.user;
}
async function refresh() {
  if (!refreshing)
    refreshing = (async () => {
      const { csrf_token } = await api<{ csrf_token: string }>(
        "/admin/auth/session-csrf",
      );
      session = await api<Session>("/admin/auth/refresh", {
        method: "POST",
        headers: { "X-CSRF-Token": csrf_token },
        body: "{}",
      });
      return session;
    })().finally(() => {
      refreshing = null;
    });
  return refreshing;
}
export async function restore() {
  try {
    return (await refresh()).user;
  } catch {
    session = null;
    return null;
  }
}
export async function logout() {
  if (session)
    await api("/admin/auth/logout", {
      method: "POST",
      headers: { "X-CSRF-Token": session.csrf_token },
      body: "{}",
    });
  session = null;
}
async function withSession<T>(request: (token: string) => Promise<T>) {
  if (!session) await refresh();
  try {
    return await request(session!.access_token);
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      await refresh();
      return request(session!.access_token);
    }
    throw error;
  }
}

export function authorized<T>(path: string, options: RequestInit = {}) {
  return withSession(token => api<T>(path, {
    ...options, headers: { ...options.headers, Authorization: `Bearer ${token}` },
  }));
}

export async function downloadAuthorized(path: string, filename: string) {
  const blob = await withSession(async token => {
    const response = await apiResponse(path, { headers: { Authorization: `Bearer ${token}` } });
    if (!response.ok) {
      const error = await response.json().catch(() => null);
      throw new ApiError(response.status, error?.error?.code || "EXPORT_UNAVAILABLE", error?.message || "Không tải được file.");
    }
    if (!response.headers.get("Content-Disposition")?.startsWith("attachment")) {
      throw new Error("Máy chủ không trả về file tải xuống.");
    }
    return response.blob();
  });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
