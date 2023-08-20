import {
  Drawer,
  FormControl,
  ImageList,
  ImageListItem,
  InputLabel,
  MenuItem,
  Paper,
  Select,
} from "@mui/material";
import Box from "@mui/material/Box";
import { useCallback, useEffect, useState } from "react";
import { OFFSET_SIZE } from "./constants";
import {
  transformPathToImagePath,
  useGetBoardsByCategory,
  useGetImagePaths,
} from "./hooks";
import { Category, isCategory } from "./types";

function rangeFromOffset(offset: number): { from: number; to: number } {
  console.log({
    from: Math.max(offset * OFFSET_SIZE - OFFSET_SIZE, 0),
    to: (offset + 1) * OFFSET_SIZE,
  });

  return {
    from: Math.max(offset * OFFSET_SIZE - OFFSET_SIZE, 0),
    to: (offset + 1) * OFFSET_SIZE,
  };
}

function App() {
  const [offset, setOffset] = useState<number>(0);
  const [category, setCategory] = useState<Category | undefined>(undefined);
  const [board, setBoard] = useState<string | undefined>(undefined);

  const boards = useGetBoardsByCategory(category);
  const imagePaths = useGetImagePaths({ category, board });
  const { from, to } = rangeFromOffset(offset);

  const onScroll = useCallback(() => {
    const bottom =
      Math.ceil(window.innerHeight + window.scrollY) >=
      document.documentElement.scrollHeight;

    if (bottom) {
      setOffset(offset + 1);
    }
  }, [offset]);

  useEffect(() => {
    window.addEventListener("scroll", onScroll);

    // Clean-up
    return () => {
      window.removeEventListener("scroll", onScroll);
    };
  }, [onScroll]);

  if (imagePaths.error || boards.error) return <>An error has occurred</>;
  if (
    imagePaths.isLoading ||
    !imagePaths.data ||
    boards.isLoading ||
    !boards.data
  )
    return <>Loading...</>;

  return (
    <>
      <Drawer anchor="top" open={true}>
        <Paper>
          <FormControl sx={{ m: 1, minWidth: 80 }}>
            <InputLabel id="select-category-label">Category</InputLabel>
            <Select
              label="Selected category"
              labelId="select-category-label"
              id="select-category"
              onChange={(event) => {
                if (typeof event.target.value !== "string")
                  throw new Error("Unexpected value from form");
                if (event.target.value === "") {
                  setCategory(undefined);
                  return;
                }
                if (!isCategory(event.target.value)) {
                  throw new Error("Unexpected category");
                }

                setCategory(event.target.value);
              }}
              value={category || ""}
            >
              <MenuItem value={""}>
                <em>Deselect</em>
              </MenuItem>
              <MenuItem value="pepe">Pepe</MenuItem>
              <MenuItem value="non-pepe">Non-Pepe</MenuItem>
              <MenuItem value="maybe">Maybe</MenuItem>
            </Select>
          </FormControl>

          <FormControl sx={{ m: 1, minWidth: 80 }}>
            <InputLabel id="select-board-label">Selected board</InputLabel>
            <Select
              label="Selected board"
              labelId="select-board-label"
              id="select-board"
              onChange={(event) => {
                if (typeof event.target.value !== "string")
                  throw new Error("Unexpected value from form");
                if (event.target.value === "") {
                  setBoard(undefined);
                  return;
                }
                setBoard(event.target.value);
              }}
              value={board || ""}
            >
              <MenuItem value={""}>
                <em>Deselect</em>
              </MenuItem>
              {boards.data.map((b) => (
                <MenuItem value={b}>{`/${b}/`}</MenuItem>
              ))}
            </Select>
          </FormControl>
        </Paper>
      </Drawer>
      <Box>
        <ImageList variant="masonry" cols={5}>
          {imagePaths.data.slice(from, to).map((s, index) => {
            return (
              <ImageListItem key={(index % 5) * 100000000 + index}>
                <img
                  src={transformPathToImagePath(s)}
                  loading="lazy"
                  style={{ width: "100%" }}
                />
              </ImageListItem>
            );
          })}
        </ImageList>
      </Box>
    </>
  );
}

export default App;
