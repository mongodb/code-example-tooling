import styles from "./Header.module.css";

import Toggle from "@leafygreen-ui/toggle";
import LogoBlock from "../logoblock/LogoBlock";

type HeaderProps = {
  darkMode: boolean;
  setDarkMode: (darkMode: boolean) => void;
};

function Header({ darkMode, setDarkMode }: HeaderProps) {
  return (
    <header className={styles.header}>
      <LogoBlock />
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
