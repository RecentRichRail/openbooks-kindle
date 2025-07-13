import React from "react";
import { createRoot } from "react-dom/client";
import { Provider } from "react-redux";
import App from "./App";
import { store } from "./state/store";

console.log("Main.tsx loading...");

const container = document.getElementById("root");
if (!container) {
  console.error("Root container not found!");
} else {
  console.log("Root container found, creating app...");
}

const root = createRoot(container!);
root.render(
  <React.StrictMode>
    <Provider store={store}>
      <App />
    </Provider>
  </React.StrictMode>
);

console.log("App rendered to root");
