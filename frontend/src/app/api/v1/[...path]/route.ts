import { NextRequest } from "next/server";
async function proxy(request: NextRequest) {
  const upstream = new URL(process.env.API_UPSTREAM_URL || "http://localhost:8080");
  upstream.pathname = request.nextUrl.pathname;
  upstream.search = request.nextUrl.search;
  const headers = new Headers();
  for (const name of ["content-type", "cookie", "authorization", "origin", "x-csrf-token", "idempotency-key", "referer", "sec-fetch-site"]) {
    const value = request.headers.get(name); if (value) headers.set(name, value);
  }
  try {
    let body: Uint8Array | undefined;
    if (!["GET", "HEAD"].includes(request.method) && request.body) {
      const reader = request.body.getReader(); const chunks: Uint8Array[] = []; let size = 0;
      while (true) { const part = await reader.read(); if (part.done) break; size += part.value.length;
        if (size > 256 * 1024) { await reader.cancel(); return Response.json({success:false,data:null,error:{code:"PAYLOAD_TOO_LARGE",fields:{}},message:"Dữ liệu vượt giới hạn."},{status:413}); }
        chunks.push(part.value);
      }
      body = new Uint8Array(size); let offset = 0; for (const chunk of chunks) { body.set(chunk, offset); offset += chunk.length; }
    }
    const response = await fetch(upstream, { method: request.method, headers, body: body as BodyInit | undefined, redirect: "manual", cache: "no-store", signal: AbortSignal.timeout(30000) });
    const output = new Headers({"Cache-Control":"no-store"});
    for (const name of ["content-type", "x-request-id", "retry-after"]) { const value = response.headers.get(name); if (value) output.set(name, value); }
    for (const cookie of response.headers.getSetCookie()) output.append("set-cookie", cookie);
    return new Response(response.body, {status:response.status, headers:output});
  } catch { return Response.json({success:false,data:null,error:{code:"SERVICE_UNAVAILABLE",fields:{}},message:"Máy chủ đang khởi động. Vui lòng thử lại."},{status:503}); }
}
export { proxy as GET, proxy as POST, proxy as PUT, proxy as PATCH, proxy as DELETE, proxy as OPTIONS };
