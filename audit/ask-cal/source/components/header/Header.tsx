import styles from "./Header.module.css";

import { H1 } from "@leafygreen-ui/typography";
import Toggle from "@leafygreen-ui/toggle";
import Logo from "@leafygreen-ui/logo";

type HeaderProps = {
  darkMode: boolean;
  setDarkMode: (darkMode: boolean) => void;
};

function Header({ darkMode, setDarkMode }: HeaderProps) {
  return (
    <header className={styles.header}>
      <div className={styles.logo_block}>
        <H1>Ask CAL</H1>
        <Logo name={"MongoDBLogoMark"} />
      </div>
      <Toggle
        aria-label="Dark mode toggle"
        checked={darkMode}
        onChange={setDarkMode}
        size="small"
      />
    </header>
  );
}

export default Header;
