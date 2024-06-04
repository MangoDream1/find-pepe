import autoBind from "auto-bind";
import sequelize, { DataTypes, Model, Op } from "sequelize";
import { Category } from "./types.js";

interface IImage extends Model {
  filePath: string;
  category: string;
  classification: number;
  href: string;
  board: string;
}

export class DB {
  sequelize: sequelize.Sequelize;
  Model: sequelize.ModelStatic<IImage>;

  constructor(deps: { sequelize: sequelize.Sequelize }) {
    this.sequelize = deps.sequelize;
    this.Model = this.sequelize.define<IImage>(
      "Images",
      {
        id: {
          type: DataTypes.INTEGER,
          autoIncrement: true,
          primaryKey: true,
        },
        filePath: {
          type: DataTypes.TEXT(),
          field: "file_path",
        },
        category: DataTypes.TEXT(),
        classification: DataTypes.INTEGER,
        href: DataTypes.TEXT(),
        board: DataTypes.TEXT(),
      },
      {
        tableName: "images",
      }
    );

    autoBind(this);
  }

  async getUniqueBoards(): Promise<string[]> {
    const result = (await this.sequelize.query(
      `SELECT DISTINCT board FROM Images`
    )) as [{ board: string }[], unknown];

    return result[0].map((r) => r.board);
  }

  async getUniqueBoardsByCategory(catagory: Category): Promise<string[]> {
    const result = (await this.sequelize.query({
      query: `SELECT DISTINCT board FROM Images WHERE category = ?`,
      values: [catagory],
    })) as [{ board: string }[], unknown];

    return result[0].map((r) => r.board);
  }

  async getImagesLocationsBySelection(
    selection: {
      boards: string[];
      categories: Category[];
    },
    options?: { offset?: number; limit?: number }
  ): Promise<string[]> {
    let where: sequelize.WhereOptions = {};
    if (selection.boards.length > 0) {
      where.board = { [Op.in]: selection.boards };
    }

    if (selection.categories.length > 0) {
      where.category = { [Op.in]: selection.categories };
    }

    return (
      await this.Model.findAll({
        where,
        offset: options?.offset,
        limit: options?.limit,
        raw: true,
        attributes: ["filePath"],
      })
    ).map((i) => i.filePath);
  }
}
