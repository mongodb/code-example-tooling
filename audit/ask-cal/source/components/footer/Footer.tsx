import styles from "./Footer.module.css";

import { Body } from "@leafygreen-ui/typography";

function Footer() {
  return (
    <div className={styles.footer}>
      <Body>This is the footer component</Body>
    </div>
  );
}

export default Footer;
