import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { AppProvider } from "./store/AppContext";
import { ErrorProvider } from "./store/ErrorContext";
import ErrorBanner from "./components/ErrorBanner";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ErrorProvider>
      <AppProvider>
        <ErrorBanner />
        <App />
      </AppProvider>
    </ErrorProvider>
  </StrictMode>,
);
