import styles from "./Viewer.module.css";

import { H2, Body } from "@leafygreen-ui/typography";

function Viewer() {
  return (
    <div className={styles.viewer}>
      <H2>Viewer Component</H2>
      <Body>This is the viewer component.</Body>
    </div>
  );
}

export default Viewer;
