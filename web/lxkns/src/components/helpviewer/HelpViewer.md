```tsx
import { MemoryRouter } from "react-router";
import { MuiMarkdown } from "components/muimarkdown";

const MyMarkdowner = (props) => (<MuiMarkdown {...props} />);

/* eslint import/no-webpack-loader-syntax: off */
import chintro from "!babel-loader!mdx-loader!./01-intro.mdx";
import chfoobar from "!babel-loader!mdx-loader!./02-foobar.mdx";
import chnew from "!babel-loader!mdx-loader!./03-newchapter.mdx";

const chapters = [
  { title: "Intro", chapter: chintro },
  { title: "Foo Bar", chapter: chfoobar },
  { title: "A New Chapter", chapter: chnew },
];

<MemoryRouter initialEntries={['/help']}>
  <HelpViewer
    chapters={chapters}
    baseroute='/help'
    style={{ height: '30ex', maxHeight: '30ex' }}
    markdowner={MyMarkdowner}
  />
</MemoryRouter>;
```
