import styles from "./Sidebar.module.css";

import { SideNav, SideNavGroup, SideNavItem } from "@leafygreen-ui/side-nav";
import Icon from "@leafygreen-ui/icon";

function Sidebar() {
  return (
    <SideNav
      widthOverride={300}
      className={styles.sidebar}
    >
      <SideNavItem>Sidebar item 1</SideNavItem>
      <SideNavItem>Sidebar item 2</SideNavItem>
      <SideNavItem>
        Sidebar item 3<SideNavItem>Sub item 1</SideNavItem>
      </SideNavItem>
      <SideNavGroup
        header="Collapsible Sidebar Group"
        collapsible
        glyph={<Icon glyph="Building" />}
      >
        <SideNavItem active>Sidebar group item 1</SideNavItem>
        <SideNavItem>Sidebar group item 2</SideNavItem>
        <SideNavGroup header="Nested Sidebar Group">
          <SideNavItem>Nested group item 1</SideNavItem>
          <SideNavItem>Nested group item 2</SideNavItem>
        </SideNavGroup>
      </SideNavGroup>
    </SideNav>
  );
}

export default Sidebar;
