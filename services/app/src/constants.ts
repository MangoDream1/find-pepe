export const API_URL =
  process.env.REACT_APP_API_URL || new URL("/api", window.location.origin);

export const IMAGE_CATEGORIES = ["maybe", "non-pepe", "pepe"] as const;
export const OFFSET_SIZE = 25;
export const SCROLL_THRESHOLD_PX = 500;
