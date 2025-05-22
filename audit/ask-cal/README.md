# Ask CAL

Welcome!

## Dev Environment

Need to use Node v20.9.0 to satisfy Leafygreen dependencies.

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
