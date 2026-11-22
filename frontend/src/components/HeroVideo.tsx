"use client";

import { useEffect, useRef, useState, useSyncExternalStore } from "react";

const motionQuery = "(prefers-reduced-motion: no-preference)";

function subscribeToMotion(onChange: () => void) {
  const media = window.matchMedia(motionQuery);
  media.addEventListener("change", onChange);
  return () => media.removeEventListener("change", onChange);
}

export default function HeroVideo() {
  const allowMotion = useSyncExternalStore(
    subscribeToMotion,
    () => window.matchMedia(motionQuery).matches,
    () => false,
  );
  return allowMotion ? <Video /> : null;
}

function Video() {
  const ref = useRef<HTMLVideoElement>(null);
  const [hasFrame, setHasFrame] = useState(false);
  const showPoster = () => setHasFrame(false);

  useEffect(() => {
    const video = ref.current;
    if (!video) return;
    let onScreen = !("IntersectionObserver" in window);
    const sync = () => {
      if (onScreen && !document.hidden) {
        void video.play().catch(() => {});
      } else {
        video.pause();
      }
    };
    const observer = onScreen ? null : new IntersectionObserver(([entry]) => {
      onScreen = entry.isIntersecting;
      sync();
    });
    observer?.observe(video);
    document.addEventListener("visibilitychange", sync);
    sync();
    return () => {
      observer?.disconnect();
      document.removeEventListener("visibilitychange", sync);
      video.pause();
    };
  }, []);

  return (
    <video
      ref={ref}
      className="home-hero-video"
      data-frame={hasFrame}
      preload="none"
      muted
      playsInline
      loop
      aria-hidden="true"
      onPlaying={() => setHasFrame(true)}
      onError={showPoster}
    >
      <source src="/images/home-coast.webm" type="video/webm" />
      <source src="/images/home-coast.mp4" type="video/mp4" onError={showPoster} />
    </video>
  );
}
