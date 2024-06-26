import autoBind from "auto-bind";
import express from "express";
import { IMAGE_RETURN_LIMIT } from "../constants.js";
import { Core } from "../core.js";
import { DB } from "../db.js";
import { Category } from "../types.js";
import {
  getBoardsByCategoryRequestSchema,
  retrieveImagesByQueryRequestSchema,
} from "./schema.js";

export class Controller {
  core: Core;
  db: DB;

  constructor(deps: { core: Core; db: DB }) {
    this.core = deps.core;
    this.db = deps.db;

    autoBind(this);
  }

  async getBoards(req: express.Request, res: express.Response) {
    const boards = await this.db.getUniqueBoards();
    res.status(200).send(boards);
  }

  async getBoardsByCategoryFromParam(
    req: express.Request,
    res: express.Response
  ) {
    const pReq = getBoardsByCategoryRequestSchema.parse(req);
    const boards = await this.db.getUniqueBoardsByCategory(
      pReq.params.category
    );
    res.status(200).send(boards);
  }

  async retrieveImagesByQuery(req: express.Request, res: express.Response) {
    const { query } = retrieveImagesByQueryRequestSchema.parse(req);
    const { category, board, offset } = query;

    const selection = {
      categories: [] as Category[],
      boards: [] as string[],
    };
    if (category) {
      selection.categories = [category];
    }
    if (board) {
      selection.boards = [board];
    }

    const images: string[] = (
      await this.db.getImagesLocationsBySelection(selection, {
        limit: IMAGE_RETURN_LIMIT,
        offset,
      })
    ).map(this.core.dbFileNameToPublicURL);

    res.status(200).send(images);
  }
}
