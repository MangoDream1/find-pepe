import path from "path";
import { DATA_DIR, PUBLIC_SERVE_LOCATION } from "./constants";

export class Core {
  private _createDataDirPath() {
    return path.join(__dirname, "..", DATA_DIR);
  }

  removeDataDirFromPath(p: string) {
    const dirPath = this._createDataDirPath();
    return path.join(PUBLIC_SERVE_LOCATION, p.replace(dirPath, ""));
  }
}
