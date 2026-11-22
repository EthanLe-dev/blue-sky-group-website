export class ApiError extends Error {
  status: number;
  code: string;
  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}
export async function apiResponse(
  path: string,
  options: RequestInit = {},
): Promise<Response> {
  if (!path.startsWith("/") || path.startsWith("//"))
    throw new Error("API path must be relative");
  return fetch("/api/v1" + path, {
    ...options,
    credentials: "include",
    cache: "no-store",
    headers: {
      ...(options.body instanceof FormData
        ? {}
        : { "Content-Type": "application/json" }),
      ...options.headers,
    },
  });
}
export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await apiResponse(path, options);
  const result = await response.json();
  if (!response.ok || !result.success)
    throw new ApiError(
      response.status,
      result.error?.code || "INTERNAL_ERROR",
      result.message || "Không thể thực hiện yêu cầu.",
    );
  return result.data as T;
}
