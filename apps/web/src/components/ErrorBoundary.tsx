import { Link, useRouteError } from "react-router-dom";

// Rendered by the router when a route element throws (e.g. a bad URL param),
// instead of the app white-screening.
export function RouteError() {
  const err = useRouteError() as Error | undefined;
  return (
    <div className="mx-auto max-w-lg px-6 py-20 text-center">
      <h1 className="font-display text-2xl font-bold">Something went wrong</h1>
      <p className="mt-2 break-words text-sm text-muted">{err?.message ?? "Unexpected error."}</p>
      <Link
        to="/"
        className="mt-5 inline-block rounded-lg border border-border px-4 py-2 text-sm hover:border-accent"
      >
        Back to dashboard
      </Link>
    </div>
  );
}

export function NotFound() {
  return (
    <div className="mx-auto max-w-lg px-6 py-20 text-center">
      <h1 className="font-display text-2xl font-bold">Page not found</h1>
      <Link
        to="/"
        className="mt-5 inline-block rounded-lg border border-border px-4 py-2 text-sm hover:border-accent"
      >
        Back to dashboard
      </Link>
    </div>
  );
}
