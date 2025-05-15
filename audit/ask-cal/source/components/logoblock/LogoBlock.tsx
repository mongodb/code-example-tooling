import styles from "./LogoBlock.module.css";

import { H1 } from "@leafygreen-ui/typography";
import Logo from "@leafygreen-ui/logo";

function LogoBlock() {
  return (
    <div className={styles.logo_block}>
      <H1>Ask CAL</H1>
      <Logo name={"MongoDBLogoMark"} />
    </div>
  );
}

export default LogoBlock;
