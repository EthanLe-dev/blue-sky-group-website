import { api, ApiError } from "./api";
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
export async function authorized<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  if (!session) await refresh();
  const run = () =>
    api<T>(path, {
      ...options,
      headers: {
        ...options.headers,
        Authorization: `Bearer ${session!.access_token}`,
      },
    });
  try {
    return await run();
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) {
      await refresh();
      return run();
    }
    throw e;
  }
}
