import styles from "./HomePage.module.css";

import { Dispatch, SetStateAction } from "react";

import { SearchInput } from "@leafygreen-ui/search-input";
import { Combobox, ComboboxOption } from "@leafygreen-ui/combobox";
import Toggle from "@leafygreen-ui/toggle";

import LogoBlock from "../../components/logoblock/LogoBlock";

interface HomepageProps {
  setIsHomepage: Dispatch<SetStateAction<boolean>>;
  setDarkMode: Dispatch<SetStateAction<boolean>>;
  darkMode: boolean;
}

function Homepage({ setIsHomepage, setDarkMode, darkMode }: HomepageProps) {
  const handleSearch = () => {
    // Handle search logic here
    console.log("Search triggered");

    setIsHomepage(false);
  };

  return (
    <div className={styles.homepage}>
      <Toggle
        aria-label="Dark mode toggle"
        checked={darkMode}
        onChange={setDarkMode}
        size="small"
        className={styles.theme_toggle}
      />
      <LogoBlock />

      <div className={styles.facet_container}>
        <Combobox
          label="Programming Language"
          placeholder="Select language"
          size="xsmall"
        >
          <ComboboxOption value="JavaScript" />
          <ComboboxOption value="C#" />
          <ComboboxOption value="Go" />
          <ComboboxOption value="Rust" />
        </Combobox>

        <Combobox
          label="Category"
          placeholder="Select category"
          size="xsmall"
        >
          <ComboboxOption value="JavaScript" />
          <ComboboxOption value="C#" />
          <ComboboxOption value="Go" />
          <ComboboxOption value="Rust" />
        </Combobox>

        <Combobox
          label="Documentation set"
          placeholder="Select docs set"
          size="xsmall"
        >
          <ComboboxOption value="JavaScript" />
          <ComboboxOption value="C#" />
          <ComboboxOption value="Go" />
          <ComboboxOption value="Rust" />
        </Combobox>
      </div>

      <SearchInput
        onSubmit={handleSearch}
        aria-label="search box"
        className={styles.search_input}
        size="large"
      ></SearchInput>
    </div>
  );
}

export default Homepage;
