import autoBind from "auto-bind";
import express from "express";
import { IMAGE_CATEGORIES } from "../constants.js";
import { Category } from "../types.js";

export class Validator {
  constructor() {
    autoBind(this);
  }

  private category =
    (category: Category | undefined) =>
    (
      req: express.Request,
      res: express.Response,
      next: express.NextFunction
    ) => {
      if (!category || !IMAGE_CATEGORIES.includes(category)) {
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
    const category = req.params.category as Category | undefined;
    this.category(category)(req, res, next);
  }

  queryCategory(
    req: express.Request,
    res: express.Response,
    next: express.NextFunction
  ) {
    const category = req.query.category as Category | undefined;
    if (!category) {
      next();
      return;
    }

    this.category(category)(req, res, next);
  }
}
