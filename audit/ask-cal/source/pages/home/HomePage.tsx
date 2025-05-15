import styles from "./HomePage.module.css";

import { PageLoader } from "@leafygreen-ui/loading-indicator";

import { useAcala } from "../../providers/UseAcala";
import Search from "../../components/search/Search";

interface HomepageProps {
  setIsHomepage: React.Dispatch<React.SetStateAction<boolean>>;
}

function Homepage({ setIsHomepage }: HomepageProps) {
  const { loading } = useAcala();

  return (
    <div className={styles.homepage}>
      {/* If loading, render BlobLoader */}
      {loading ? (
        <div className={styles.loading_container}>
          <PageLoader description="Looking for code examples..." />
        </div>
      ) : (
        <Search
          isHomepage={true}
          setIsHomepage={setIsHomepage}
        />
      )}
    </div>
  );
}

export default Homepage;
