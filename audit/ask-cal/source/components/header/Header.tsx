import styles from "./Header.module.css";

import LogoBlock from "../logoblock/LogoBlock";
import Search from "../search/Search";

type HeaderProps = {
  isHomepage: boolean;
  setIsHomepage: React.Dispatch<React.SetStateAction<boolean>>;
};

function Header({ isHomepage, setIsHomepage }: HeaderProps) {
  return (
    <header className={isHomepage ? styles.header : styles.header_results}>
      <Logo />

      {!isHomepage && (
        <Search
          isHomepage={isHomepage}
          setIsHomepage={setIsHomepage}
        />
      )}
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
