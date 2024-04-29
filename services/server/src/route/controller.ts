import autoBind from "auto-bind";
import express from "express";
import { Core } from "../core.js";
import { DB } from "../db.js";
import { Category } from "../types.js";

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
    const boards = await this.db.getUniqueBoardsByCategory(
      req.params.category as Category
    );
    res.status(200).send(boards);
  }

  // TODO: add skip / limit
  async retrieveImagesByQuery(req: express.Request, res: express.Response) {
    const category = req.query.category as Category | undefined;
    const board = req.query.board as string | undefined;

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
      await this.db.getImagesLocationsBySelection(selection)
    ).map(this.core.dbFileNameToPublicURL);

    res.status(200).send(images);
  }
}
