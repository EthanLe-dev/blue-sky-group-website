import assert from "node:assert/strict";
import test from "node:test";
import { gsap } from "gsap";

test("GSAP can build and seek a timeline for browser animation", () => {
  const subject = { opacity: 0 };
  const timeline = gsap.timeline({ paused: true }).to(subject, {
    opacity: 1,
    duration: 0.4,
  });

  timeline.progress(1);
  assert.equal(subject.opacity, 1);
  timeline.kill();
});
