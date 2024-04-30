import { z } from "zod";
import { IMAGE_CATEGORIES } from "../constants.js";

export const getBoardsByCategoryRequestSchema = z
  .object({ params: z.object({ category: z.enum(IMAGE_CATEGORIES) }) })
  .passthrough();

export const retrieveImagesByQueryRequestSchema = z
  .object({
    query: z.object({
      category: z.enum(IMAGE_CATEGORIES).optional(),
      board: z.string().optional(),
      offset: z.coerce.number().gte(0).optional(),
    }),
  })
  .passthrough();
