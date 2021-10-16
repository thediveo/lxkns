### Synchronous Import

In this example we (synchronously) import markdown "compiled" into
[MDX](https://mdxjs.com) and feed that into our `MuiMarkdown` component. Please
note that in case of Typescript we need to tell eslint to bugger off:

```tsx static
/* eslint import/no-webpack-loader-syntax: off */
import ExampleMDX from "!babel-loader!mdx-loader!./example/example.mdx";
```

```tsx
import ExampleMDX from "!babel-loader!mdx-loader!./example/example.mdx";

<MuiMarkdown mdx={ExampleMDX} />;
```

### Lazy Import

Lazy loading of potentially larger markdown/MDX documents is also supported and
works as follows:

```tsx static
const LazyExampleMDX = React.lazy(
  () => import("!babel-loader!mdx-loader!./example/minexample.mdx")
);
```

### Light/Dark Themes Support

This component has basic support for light and dark theming.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import MinExampleMDX from "!babel-loader!mdx-loader!./example/minexample.mdx";
import { createTheme, ThemeProvider, makeStyles } from "@mui/material";

const themes = [
  createTheme({ palette: { mode: "light" } }),
  createTheme({ palette: { mode: "dark" } }),
];

<>
  {themes.map((theme, idx) => (
    <ThemeProvider key={idx} theme={theme}>
      {idx > 0 && <br/>}
      with {theme.palette.type} theme:
      <ComponentCard>
        <MuiMarkdown mdx={MinExampleMDX} />
      </ComponentCard>
    </ThemeProvider>
  ))}
</>;
```
