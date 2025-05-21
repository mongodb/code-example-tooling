import styles from "./HomePage.module.css";

import { PageLoader } from "@leafygreen-ui/loading-indicator";
import { Body } from "@leafygreen-ui/typography";

import { useSearch } from "../../providers/Hooks";
import Search from "../../components/search/Search";

interface HomepageProps {
  setIsHomepage: React.Dispatch<React.SetStateAction<boolean>>;
}

function Homepage({ setIsHomepage }: HomepageProps) {
  const { loadingSearch } = useSearch();

  return (
    <div className={styles.homepage}>
      <div className={styles.description}>
        <Body baseFontSize={16}>
          Welcome to your hub for MongoDB code examples! Easily search curated
          examples from our documentation and get instant explanations from the
          Docs AI Chatbot for guidance.
        </Body>
      </div>

      {loadingSearch ? (
        <div className={styles.loading_container}>
          <PageLoader description="Looking for code examples..." />
        </div>
      ) : (
        <Search
          isHomepage={true}
          setIsHomepage={setIsHomepage}
        />
      )}

      <div className={styles.background_image}></div>
    </div>
  );
}

export default Homepage;
