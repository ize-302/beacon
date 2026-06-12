/* @refresh reload */
import { render } from "solid-js/web";
import "./style.css";
import { Route, Router } from "@solidjs/router";
import Home from "./pages/Home";
import NotFound from "./pages/NotFound";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";

const App = (props: any) => <>{props.children}</>;

const root = document.getElementById("root");
const client = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: Infinity,
      refetchOnWindowFocus: false,
    },
  },
});

render(
  () => (
    <QueryClientProvider client={client}>
      <Router root={App}>
        <Route path="/" component={Home} />
        <Route path="*paramName" component={NotFound} />
      </Router>
    </QueryClientProvider>
  ),
  root!,
);
