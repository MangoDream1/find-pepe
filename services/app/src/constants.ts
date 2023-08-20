export const API_URL =
  window.REACT_APP_API_URL ||
  process.env.REACT_APP_API_URL ||
  "http://localhost:5000/";

export const IMAGE_CATEGORIES = ["maybe", "non-pepe", "pepe"] as const;
export const OFFSET_SIZE = 50;
