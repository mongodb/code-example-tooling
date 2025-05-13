# React + TypeScript + Vite

This template provides a minimal setup to get React working in Vite with HMR and some ESLint rules.

Currently, two official plugins are available:

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) uses [Babel](https://babeljs.io/) for Fast Refresh
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) uses [SWC](https://swc.rs/) for Fast Refresh

## Expanding the ESLint configuration

If you are developing a production application, we recommend updating the configuration to enable type-aware lint rules:

```js
export default tseslint.config({
  extends: [
    // Remove ...tseslint.configs.recommended and replace with this
    ...tseslint.configs.recommendedTypeChecked,
    // Alternatively, use this for stricter rules
    ...tseslint.configs.strictTypeChecked,
    // Optionally, add this for stylistic rules
    ...tseslint.configs.stylisticTypeChecked,
  ],
  languageOptions: {
    // other options...
    parserOptions: {
      project: ["./tsconfig.node.json", "./tsconfig.app.json"],
      tsconfigRootDir: import.meta.dirname,
    },
  },
});
```

You can also install [eslint-plugin-react-x](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-x) and [eslint-plugin-react-dom](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-dom) for React-specific lint rules:

```js
// eslint.config.js
import reactX from "eslint-plugin-react-x";
import reactDom from "eslint-plugin-react-dom";

export default tseslint.config({
  plugins: {
    // Add the react-x and react-dom plugins
    "react-x": reactX,
    "react-dom": reactDom,
  },
  rules: {
    // other rules...
    // Enable its recommended typescript rules
    ...reactX.configs["recommended-typescript"].rules,
    ...reactDom.configs.recommended.rules,
  },
});
```

## Testing

We use Vitest and React Testing Library in our testing framework.

- [Vitest docs](https://vitest.dev/guide/)
- [React Testing Library docs](https://testing-library.com/docs/react-testing-library/intro/)

Notably, Vitest offers [JetBrains and VS Code extensions](https://vitest.dev/guide/ide.html).

### Unit tests

Unit test files live alongside the file that they're testing. For example, a search component named `Search.tsx` should have a test file alongside it named `Search.test.tsx`.

```
/
├─ source/
│  ├─ components/
│  │  ├─ Search.tsx
│  │  ├─ Search.test.tsx
```

### End-to-end tests

End-to-end tests go in the `tests` directory and should be modeled on specific user flows. An end-to-end test should test the entire app and include the full context of the app.

```
/
├─ source/
│  ├─ tests/
│  │  ├─ SearchForCodeExample.test.tsx
│  │  ├─ RequestNewCodeExample.test.tsx
│  │  ├─ LeaveSearchFeedback.test.tsx
│  │  ├─ GetExampleContextFromLlm.test.tsx
```
