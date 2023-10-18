import { Category } from "./types";
import sequelize from "sequelize";

// https://sequelize.org/docs/v6/core-concepts/model-basics/

export class DB {
  sequelize: sequelize.Sequelize;

  constructor(deps: { sequelize: sequelize.Sequelize }) {
    this.sequelize = deps.sequelize;
  }

  async getUniqueBoards(): Promise<string[]> {

    this.sequelize.

    return [];
  }

  async getUniqueBoardsByCategory(catagory: Category): Promise<string[]> {
    return [];
  }

  async *getImagesBySelection(
    selection: {
      boards: string[];
      categories: Category[];
    },
    options?: { skip: number }
  ): AsyncGenerator<string, undefined> {}
}
