import assert from "node:assert/strict";
import test from "node:test";
import { homeCatalog } from "../src/lib/home-catalog.ts";

test("home catalog keeps editorial order and caps tour cards at six", () => {
  const tours = Array.from({ length: 8 }, (_, id) => ({ id: String(id) }));
  assert.deepEqual(homeCatalog({ featured_tours: tours }).tours, tours.slice(0, 6));
  assert.equal(homeCatalog({ featured_tours: tours }).state, "ready");
});

test("a single tour stays visible without banners", () => {
  const tour = { id: "1" };
  const home = homeCatalog({ featured_tours: [tour] });
  assert.deepEqual(home.tours, [tour]);
  assert.deepEqual(home.banners, []);
  assert.equal(home.state, "ready");
});

test("an empty collection and a failed request produce separate states", () => {
  assert.equal(homeCatalog({ featured_tours: [] }).state, "empty");
  assert.equal(homeCatalog(null).state, "error");
  assert.deepEqual(homeCatalog(null).banners, []);
});
