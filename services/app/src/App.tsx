import MenuIcon from "@mui/icons-material/Menu";
import {
  Drawer,
  FormControl,
  Grid,
  IconButton,
  ImageList,
  ImageListItem,
  InputLabel,
  MenuItem,
  Select,
} from "@mui/material";
import Box from "@mui/material/Box";
import { ReactNode, useCallback, useEffect, useMemo, useState } from "react";
import { OFFSET_SIZE, SCROLL_THRESHOLD_PX } from "./constants";
import {
  transformPathToImagePath,
  useGetBoardsByCategory,
  useGetImagePaths,
} from "./hooks";
import { Category, isCategory } from "./types";

function App() {
  const [offsetIndex, setIndexOffset] = useState<number>(0);
  const [category, setCategory] = useState<Category | undefined>(undefined);
  const [board, setBoard] = useState<string | undefined>(undefined);
  const [showMenu, setShowMenu] = useState<boolean>(false);

  const onScroll = useCallback(() => {
    const bottom =
      Math.ceil(window.innerHeight + window.scrollY) + SCROLL_THRESHOLD_PX >=
      document.documentElement.scrollHeight;

    if (bottom) {
      setIndexOffset(offsetIndex + 1);
    }
  }, [offsetIndex]);

  useEffect(() => {
    setIndexOffset(0);
  }, [board, category]);

  const offset = useMemo(() => {
    return Math.max(offsetIndex * OFFSET_SIZE - OFFSET_SIZE, 0);
  }, [offsetIndex]);

  const boards = useGetBoardsByCategory(category);
  const imagePaths = useGetImagePaths({ category, board, offset: offset });

  useEffect(() => {
    if (!imagePaths.isLoading) {
      window.addEventListener("scroll", onScroll);
    }

    // Clean-up
    return () => {
      window.removeEventListener("scroll", onScroll);
    };
  }, [onScroll, imagePaths.isLoading]);

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
      <MenuDrawer showMenu={showMenu} setShowMenu={setShowMenu}>
        <MenuCategorySelect
          selectedCategory={category}
          setCategory={setCategory}
        />
        <MenuBoardSelect
          selectedBoard={board}
          boards={boards.data}
          setBoard={setBoard}
        />
      </MenuDrawer>
      <Images imagePaths={imagePaths.data} />
    </>
  );
}

function Images(props: { imagePaths: string[] }) {
  return (
    <Box>
      <ImageList variant="masonry" cols={5}>
        {props.imagePaths.map((s, index) => {
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
  );
}

function MenuDrawer(props: {
  showMenu: boolean;
  setShowMenu: (v: boolean) => void;
  children: ReactNode[];
}) {
  return (
    <>
      <IconButton
        sx={{
          position: "fixed",
          top: 15,
          right: 15,
          "z-index": 10000000,
          "background-color": "white",
          "border-top-left-radius": "50% 20px",
          "border-top-right-radius": "50% 20px",
        }}
        onClick={() => props.setShowMenu(!props.showMenu)}
      >
        <MenuIcon />
      </IconButton>

      <Drawer
        anchor="top"
        open={props.showMenu}
        ModalProps={{ onBackdropClick: () => props.setShowMenu(false) }}
      >
        <Grid
          container
          spacing={0}
          direction="row"
          alignItems="center"
          justifyContent="center"
        >
          {props.children}
        </Grid>
      </Drawer>
    </>
  );
}

function MenuCategorySelect(props: {
  selectedCategory: Category | undefined;
  setCategory: (category: Category | undefined) => void;
}) {
  return (
    <FormControl sx={{ m: 1, minWidth: 150 }}>
      <InputLabel id="select-category-label">Category</InputLabel>
      <Select
        label="Selected category"
        labelId="select-category-label"
        id="select-category"
        onChange={(event) => {
          if (typeof event.target.value !== "string")
            throw new Error("Unexpected value from form");
          if (event.target.value === "") {
            props.setCategory(undefined);
            return;
          }
          if (!isCategory(event.target.value)) {
            throw new Error("Unexpected category");
          }

          props.setCategory(event.target.value);
        }}
        value={props.selectedCategory || ""}
      >
        <MenuItem value={""}>
          <em>Deselect</em>
        </MenuItem>
        <MenuItem value="pepe">Pepe</MenuItem>
        <MenuItem value="non-pepe">Non-Pepe</MenuItem>
        <MenuItem value="maybe">Maybe</MenuItem>
      </Select>
    </FormControl>
  );
}

function MenuBoardSelect(props: {
  selectedBoard: string | undefined;
  boards: string[];
  setBoard: (board: string | undefined) => void;
}) {
  return (
    <FormControl sx={{ m: 1, minWidth: 150 }}>
      <InputLabel id="select-board-label">Selected board</InputLabel>
      <Select
        label="Selected board"
        labelId="select-board-label"
        id="select-board"
        onChange={(event) => {
          if (typeof event.target.value !== "string")
            throw new Error("Unexpected value from form");
          if (event.target.value === "") {
            props.setBoard(undefined);
            return;
          }
          props.setBoard(event.target.value);
        }}
        value={props.selectedBoard || ""}
      >
        <MenuItem key={"empty"} value={""}>
          <em>Deselect</em>
        </MenuItem>
        {props.boards.map((b) => (
          <MenuItem key={b} value={b}>{`/${b}/`}</MenuItem>
        ))}
      </Select>
    </FormControl>
  );
}

export default App;
