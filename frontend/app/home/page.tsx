import { fetchPosts, fetchPopularPosts } from "../_action/posts.action";
import HomePageClient from "./home-client";

export default function HomePage() {
  return <HomePageClient fetchPosts={fetchPosts} fetchPopularPosts={fetchPopularPosts} />;
}