import assert from "node:assert/strict";
import test from "node:test";
import { resolveContact, companyContact } from "../src/lib/contact.ts";
import { setAnalytics, contactEvent, selectTourEvent } from "../src/lib/analytics.ts";

test("contact settings use approved fallback and verified Admin configuration", () => {
  assert.deepEqual(resolveContact(), companyContact);
  assert.equal(resolveContact({ is_demo: true, hotline: "0900000000" }).phoneHref, companyContact.phoneHref);
  const real = resolveContact({ is_demo: false, hotline: "+84 989 123 456", zalo_url: "https://zalo.me/84989123456" });
  assert.equal(real.phoneHref, "tel:+84989123456");
  assert.equal(real.zaloHref, "https://zalo.me/84989123456");
  assert.equal(resolveContact({ is_demo: false, hotline: "bad", zalo_url: "javascript:alert(1)" }).zaloHref, companyContact.zaloHref);
  assert.equal(resolveContact({ is_demo: false, zalo_url: "https://zalo.me.attacker.example/" }).zaloHref, companyContact.zaloHref);
});

test("conversion events require consent and allowlist fields without personal data", () => {
  const tour = "00000000-0000-0000-0000-000000000001";
  const descriptors = ["window", "document", "location"].map(key => [key, Object.getOwnPropertyDescriptor(globalThis, key)]);
  try {
    globalThis.window = {};
    globalThis.document = { createElement: () => ({}), head: { append() {} } };
    globalThis.location = { origin: "https://blue-sky.example", pathname: "/contact" };
    setAnalytics("G-TEST123", false, false);
    contactEvent("zalo", "hero", tour);
    selectTourEvent(tour, "home");
    assert.equal(window.dataLayer, undefined);
    setAnalytics("G-TEST123", true, false);
    contactEvent("zalo", "hero", tour);
    selectTourEvent(tour, "home");
    contactEvent("zalo", "tour_card", "customer@example.com");
    const events = window.dataLayer.map(args => Array.from(args));
    assert.equal(events.filter(e => e[1] === "contact_click").length, 2);
    assert.equal(events.filter(e => e[1] === "select_item").length, 1);
    assert.equal(JSON.stringify(events).includes("customer@example.com"), false);
    const size = window.dataLayer.length;
    contactEvent("email@example.com", "hero", tour);
    contactEvent("zalo", "phone=0989123456", tour);
    selectTourEvent("invalid", "home");
    selectTourEvent(tour, "customer@example.com");
    assert.equal(window.dataLayer.length, size);
    setAnalytics("G-TEST123", false, false);
    contactEvent("phone", "header", tour);
    selectTourEvent(tour, "catalog");
    assert.equal(window.dataLayer.length, size);
  } finally {
    for (const [key, descriptor] of descriptors) {
      if (descriptor) Object.defineProperty(globalThis, key, descriptor);
      else delete globalThis[key];
    }
  }
});
