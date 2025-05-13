import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import App from "./App";

describe("App", () => {
  test("renders without crashing", async () => {
    // render App component
    render(<App />);

    // check if the heading is in the document
    const heading = await screen.findByRole("heading", {
      name: /Vite \+ React \+ LeafyGreen/i,
    });
    expect(heading).toBeInTheDocument();

    console.log("Heading found:", heading.innerHTML);
  });
});
