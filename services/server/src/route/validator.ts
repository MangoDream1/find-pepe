import express from "express";
import { Category } from "../types";
import { IMAGE_CATEGORIES } from "../constants";

export class Validator {
  constructor() {}

  private category =
    (category: Category) =>
    (
      req: express.Request,
      res: express.Response,
      next: express.NextFunction
    ) => {
      if (!IMAGE_CATEGORIES.includes(category)) {
        res
          .status(400)
          .send(`Category should be one of: [${IMAGE_CATEGORIES.join(", ")}]`);
        return;
      }

      next();
    };

  paramCategory(
    req: express.Request,
    res: express.Response,
    next: express.NextFunction
  ) {
    this.category(req.params.category as Category)(req, res, next);
  }

  queryCategory(
    req: express.Request,
    res: express.Response,
    next: express.NextFunction
  ) {
    this.category(req.query.category as Category)(req, res, next);
  }
}
