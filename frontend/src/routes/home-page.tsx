import { Link } from "react-router";

export function HomePage() {
  return (
    <main className="mx-auto max-w-2xl p-8">
      <h1 className="text-3xl font-bold">go-react-template</h1>
      <p className="mt-4 text-gray-700">This template has a Go API and a React frontend.</p>
      <p className="mt-4">
        <Link className="text-blue-800 underline" to="/register">
          Register
        </Link>
      </p>
    </main>
  );
}
