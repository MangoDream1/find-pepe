import { ImageList, ImageListItem } from "@mui/material";
import Box from "@mui/material/Box";
import { useCallback, useEffect, useState } from "react";
import { OFFSET_SIZE } from "./constants";
import { transformPathToImagePath, useGetImagePaths } from "./hooks";

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

  const response = useGetImagePaths();
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

  if (response.error) return <>An error has occurred</>;
  if (response.isLoading || !response.data) return <>Loading...</>;

  return (
    <Box>
      <ImageList variant="masonry" cols={5}>
        {response.data.slice(from, to).map((s, index) => (
          <ImageListItem key={index}>
            <img
              src={transformPathToImagePath(s)}
              loading="lazy"
              style={{ width: "100%" }}
            />
          </ImageListItem>
        ))}
      </ImageList>
    </Box>
  );
}

export default App;
