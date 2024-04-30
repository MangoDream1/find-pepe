import express from "express";
import { Core } from "../core.js";
import { DB } from "../db.js";
import { Controller } from "./controller.js";
import { Validator } from "./validator.js";

export const newRouter = (deps: { core: Core; db: DB }) => {
  const router = express.Router();

  const controller = new Controller(deps);
  const validator = new Validator();

  router.get("/boards", controller.getBoards);
  router.get(
    "/boards/:category",
    validator.paramCategory,
    controller.getBoardsByCategoryFromParam
  );
  router.get(
    "/",
    validator.queryCategory,
    validator.queryOffset,
    controller.retrieveImagesByQuery
  );

  return router;
};
