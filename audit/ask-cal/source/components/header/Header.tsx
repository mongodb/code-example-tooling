import styles from "./Header.module.css";

import LogoBlock from "../logoblock/LogoBlock";
import Search from "../search/Search";

type HeaderProps = {
  isHomepage: boolean;
};

function Header({ isHomepage }: HeaderProps) {
  return (
    <header className={isHomepage ? styles.header : styles.header_results}>
      <Logo />

      {!isHomepage && <Search isHomepage={isHomepage} />}
    </header>
  );
}

function Logo() {
  return (
    <div>
      <LogoBlock />
    </div>
  );
}

export default Header;
