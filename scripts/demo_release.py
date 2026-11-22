"""Prepare and verify the demo release without storing credentials in the checkout."""
import argparse
import json
import os
from pathlib import Path
import re
import time
from urllib.error import HTTPError, URLError
from urllib.parse import urlsplit
from urllib.request import HTTPRedirectHandler, Request, build_opener


class NoRedirect(HTTPRedirectHandler):
    def redirect_request(self, *args):
        return None


def origin(value):
    url = urlsplit(value)
    if not re.fullmatch(r"https://[a-zA-Z0-9.-]+", value) or not url.hostname:
        raise ValueError("expected an HTTPS origin without path or credentials")
    return value


def revision(value):
    if not re.fullmatch(r"[0-9a-f]{40}", value):
        raise ValueError("expected a Git SHA")
    return value


def prepare(root, sha, api_url):
    revision(sha)
    origin(api_url)
    for app in ("frontend", "admin"):
        public = Path(root) / app / "public"
        public.mkdir(parents=True, exist_ok=True)
        (public / "revision.txt").write_text(sha + "\n", encoding="ascii")
    config = {
        "framework": "vite",
        "rewrites": [{"source": "/api/v1/:path*", "destination": api_url + "/api/v1/:path*"}],
        "headers": [{
            "source": "/(.*)",
            "headers": [{"key": "X-Robots-Tag", "value": "noindex, nofollow"}],
        }],
    }
    (Path(root) / "admin" / "vercel.json").write_text(json.dumps(config, indent=2) + "\n", encoding="utf-8")


def request(url, *, token=None, data=None, headers=None):
    outbound = dict(headers or {})
    if token:
        outbound["Authorization"] = "Bearer " + token
    if data is not None:
        outbound["Content-Type"] = "application/json"
    req = Request(url, data=None if data is None else json.dumps(data).encode(), headers=outbound)
    try:
        with build_opener(NoRedirect()).open(req, timeout=25) as resp:
            return resp.status, resp.headers, resp.read(2 * 1024 * 1024)
    except HTTPError as error:
        return error.code, error.headers, b""


def render_api(service, token):
    if not re.fullmatch(r"srv-[a-zA-Z0-9]+", service):
        raise ValueError("invalid Render service ID")
    endpoint = "https://api.render.com/v1/services/" + service + "/deploys"

    def get(path, data=None):
        status, _, body = request(endpoint + path, token=token, data=data)
        if status not in (200, 201, 202):
            raise RuntimeError(f"Render API returned HTTP {status}")
        return json.loads(body)
    return get


def check_render(sha, deploy):
    reported = (deploy.get("commit") or {}).get("id")
    if reported and reported != sha:
        raise RuntimeError("Render deployed a different Git SHA")
    status = deploy.get("status")
    if status == "live" and reported != sha:
        raise RuntimeError("Render did not confirm the live Git SHA")
    if status in {"build_failed", "update_failed", "pre_deploy_failed", "canceled", "deactivated"}:
        raise RuntimeError(f"Render deployment ended with {status}")
    return status == "live"


def render(sha, service, token, api_url):
    get = render_api(service, token)
    deployment = get("", {"commitId": revision(sha)})
    deployment_id = deployment.get("id", "")
    if not re.fullmatch(r"dep-[a-zA-Z0-9]+", deployment_id):
        raise RuntimeError("Render did not return a deployment ID")
    for _ in range(84):
        if check_render(sha, get("/" + deployment_id)):
            assert_ready(api_url, sha)
            print(f"Render live: {deployment_id}")
            return
        time.sleep(10)
    raise RuntimeError("timed out waiting for Render")


def assert_ready(api_url, sha):
    code, headers, body = request(origin(api_url) + "/readyz")
    if code != 200 or headers.get("X-App-Revision") != sha or json.loads(body).get("status") != "ready":
        raise RuntimeError("API is not ready on the tested Git SHA")


def smoke(api_url, public_url, admin_url, sha):
    assert_ready(api_url, sha)
    for label, url in (("public", public_url), ("admin", admin_url)):
        base = origin(url)
        for _ in range(12):
            status, _, body = request(base + "/revision.txt")
            if status == 200 and body.decode("ascii", errors="replace").strip() == sha:
                break
            time.sleep(5)
        else:
            raise RuntimeError(f"{label} does not serve the expected Git SHA")
        status, _, body = request(base + "/")
        if status != 200 or b"Blue Sky Group".lower() not in body.lower():
            raise RuntimeError(f"{label} home page unavailable")

    code, _, body = request(public_url + "/api/v1/tours?page_size=1")
    if code != 200 or not json.loads(body).get("data"):
        raise RuntimeError("public frontend cannot list seeded tours")
    code, headers, body = request(public_url + "/api/v1/guest-sessions", data={}, headers={"Origin": public_url})
    cookies = " ".join(headers.get_all("Set-Cookie", []))
    if code not in (200, 201) or "Secure" not in cookies or "HttpOnly" not in cookies or not json.loads(body)["data"].get("csrf_token"):
        raise RuntimeError("public proxy did not establish a secure guest session")
    code, _, _ = request(public_url + "/api/v1/guest-sessions", data={}, headers={"Origin": "https://invalid.example"})
    if code != 403:
        raise RuntimeError("API origin check failed")
    code, headers, _ = request(admin_url + "/api/v1/admin/auth/me")
    if code != 401 or "no-store" not in headers.get("Cache-Control", ""):
        raise RuntimeError("Admin proxy did not retain the private API status or cache control")
    print("Backend, storefront and admin smoke checks passed.")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("action", choices=["prepare", "render", "smoke"])
    parser.add_argument("--root")
    args = parser.parse_args()
    sha = revision(os.environ["RELEASE_SHA"])
    api_url = origin(os.environ["RENDER_URL"])
    try:
        if args.action == "prepare":
            prepare(args.root, sha, api_url)
        elif args.action == "render":
            render(sha, os.environ["RENDER_SERVICE_ID"], os.environ["RENDER_API_KEY"], api_url)
        else:
            smoke(api_url, os.environ["PUBLIC_DEMO_URL"], os.environ["ADMIN_DEMO_URL"], sha)
    except URLError:
        raise SystemExit("Network request failed during demo release.") from None


if __name__ == "__main__":
    main()
