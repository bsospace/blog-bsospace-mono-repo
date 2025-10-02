import { Metadata } from "next";
import NotFound from "./components/not-found";

export const metadata: Metadata = {
  title: "Page Not Found - 404",
  description: "The page you are looking for could not be found.",
  robots: {
    index: false,
    follow: false,
  },
};

export default function NotFoundPage() {
    return <NotFound />;
}