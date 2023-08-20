import { IMAGE_CATEGORIES } from "./constants";

export type Category = (typeof IMAGE_CATEGORIES)[number];

export function isCategory(unknown: unknown): unknown is Category {
  return IMAGE_CATEGORIES.includes(unknown as Category);
}
