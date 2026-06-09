/* @refresh reload */
import { render } from "solid-js/web";
import "./style.css";
import { Route, Router } from "@solidjs/router";
import Home from "./pages/Home";
import NotFound from "./pages/NotFound";

const App = (props: any) => <>{props.children}</>;

const root = document.getElementById("root");

render(
  () => (
    <Router root={App}>
      <Route path="/" component={Home} />
      <Route path="*paramName" component={NotFound} />
    </Router>
  ),
  root!,
);
