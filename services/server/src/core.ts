import autoBind from "auto-bind";
import path from "path";
import { PUBLIC_SERVE_LOCATION } from "./constants.js";

export class Core {
  constructor() {
    autoBind(this);
  }

  dbFileNameToPublicURL(p: string) {
    const fileName = path.basename(p);
    const filePath = path.join(PUBLIC_SERVE_LOCATION, fileName);

    return filePath;
  }
}
