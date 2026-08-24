import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { App } from "./App";
import { Dashboard } from "./pages/Dashboard";
import { MonitorDetail } from "./pages/MonitorDetail";
import { LoadTest } from "./pages/LoadTest";
import { SslInspector } from "./pages/SslInspector";
import { Settings } from "./pages/Settings";
import { applyTheme, useUI } from "./lib/store";
import "./index.css";

applyTheme(useUI.getState().theme);

const queryClient = new QueryClient();

const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    children: [
      { index: true, element: <Dashboard /> },
      { path: "monitors/:url", element: <MonitorDetail /> },
      { path: "loadtest", element: <LoadTest /> },
      { path: "ssl", element: <SslInspector /> },
      { path: "settings", element: <Settings /> },
    ],
  },
]);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
);
