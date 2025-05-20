import styles from "./LogoBlock.module.css";

import { H1 } from "@leafygreen-ui/typography";
import Logo from "@leafygreen-ui/logo";

function LogoBlock() {
  return (
    <div className={styles.logo_block}>
      <Logo name={"MongoDBLogoMark"} />
      <H1>Ask CAL</H1>
    </div>
  );
}

export default LogoBlock;
