import express from "express";
import { ZodSchema } from "zod";
import {
  getBoardsByCategoryRequestSchema,
  retrieveImagesByQueryRequestSchema,
} from "./schema.js";

function zodToValidator(
  requestSchema: ZodSchema
): (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction
) => void {
  return (req, res, next) => {
    const result = requestSchema.safeParse(req);
    if (!result.success) {
      res.status(400).send({
        message: "Bad request",
        error: result.error.flatten(),
      });
      return;
    }

    next();
  };
}

export const getBoardsByCategoryValidator = zodToValidator(
  getBoardsByCategoryRequestSchema
);

export const retrieveImagesByQueryRequest = zodToValidator(
  retrieveImagesByQueryRequestSchema
);
