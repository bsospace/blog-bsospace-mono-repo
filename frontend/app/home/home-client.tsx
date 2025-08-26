/* eslint-disable react-hooks/exhaustive-deps */
"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { Meta } from "../interfaces";
import BlogCard from "../components/Blog";
import { Post } from "../interfaces";
import { FiCode, FiCpu, FiSearch, FiTrendingUp, FiZap } from "react-icons/fi";
import Loading from "../components/Loading";
import { CookiesConsentModal } from "../components/cookies-consent-modal";

export default function HomePageClient({ fetchPosts }: { fetchPosts: (page: number, limit: number, search: string) => Promise<{ data: Post[]; meta: Meta }> }) {
  // Constants
  const NO_SEARCH_QUERY = "";
  
  const [posts, setPosts] = useState<Post[]>([]);
  const [popularPosts, setPopularPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingPopular, setLoadingPopular] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [hasNextPage, setHasNextPage] = useState(false);
  const [isFirstLoad, setIsFirstLoad] = useState(true);

  const observerRef = useRef<HTMLDivElement | null>(null);
  const searchTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Improved debounced search
  const debouncedSearch = useCallback((query: string) => {
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }

    searchTimeoutRef.current = setTimeout(() => {
      setSearchQuery(query);
      setPage(1);
      setPosts([]);
    }, 300);
  }, []);

  const getPosts = useCallback(async (pageNum = 1, isLoadMore = false) => {
    try {
      if (pageNum === 1 && !isLoadMore) {
        setLoading(true);
      } else {
        setLoadingMore(true);
      }

      const res = await fetchPosts(pageNum, 5, searchQuery);

      if (pageNum === 1) {
        setPosts(res.data);
      } else {
        setPosts(prevPosts => [...prevPosts, ...res.data]);
      }

      setHasNextPage(res.meta.hasNextPage);
      setPage(pageNum);

    } catch (err) {
      console.error("Failed to load posts:", err);
    } finally {
      setLoading(false);
      setLoadingMore(false);
      setIsFirstLoad(false);
    }
  }, [searchQuery, fetchPosts]);

  const getPopularPosts = useCallback(async () => {
    try {
      setLoadingPopular(true);
      const res = await fetchPosts(1, 5, NO_SEARCH_QUERY);
      setPopularPosts(res.data);
    } catch (err) {
      console.error("Failed to load popular posts:", err);
    } finally {
      setLoadingPopular(false);
    }
  }, []);

  // Improved Intersection Observer
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        const [entry] = entries;
        if (
          entry.isIntersecting &&
          hasNextPage &&
          !loadingMore &&
          !loading &&
          posts?.length > 0
        ) {
          getPosts(page + 1, true);
        }
      },
      {
        threshold: 0.1,
        rootMargin: '100px'
      }
    );

    if (observerRef.current) {
      observer.observe(observerRef.current);
    }

    return () => {
      if (observerRef.current) {
        observer.unobserve(observerRef.current);
      }
    };
  }, [hasNextPage, loadingMore, loading, posts?.length, page, getPosts]);

  // Initial load and search effect
  useEffect(() => {
    if (isFirstLoad) {
      getPosts(1);
      getPopularPosts();
    } else {
      getPosts(1);
    }
  }, [searchQuery]);

  // Handle search input
  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    debouncedSearch(value);
  };

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
    };
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950">
      <div className="container mx-auto px-4 py-6">
        {/* Header Section - More Compact */}
        <div className="absolute inset-0 overflow-hidden mt-16">
          <div className="absolute top-10 left-1/4 w-48 h-48 bg-gradient-to-br from-orange-500/10 to-red-500/10 rounded-full blur-3xl"></div>
          <div className="absolute top-20 right-1/4 w-40 h-40 bg-gradient-to-br from-red-500/10 to-yellow-500/10 rounded-full blur-3xl"></div>
        </div>

        <header className="text-center mb-12 relative">
          <div className="relative z-10">
            <div className="flex justify-center items-center mb-4">
              <FiCode className="w-6 h-6 text-orange-400 mr-2" />
              <h1 className="text-2xl md:text-5xl font-bold" style={{
                background: 'linear-gradient(to right, #fb923c, #f87171, #facc15)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                backgroundClip: 'text',
                lineHeight: '1.2'
              }}>
                Be Simple but Outstanding
              </h1>
              <FiCpu className="w-6 h-6 text-red-400 ml-2" />
            </div>

            {/* Compact search bar */}
            <div className="relative max-w-xl mx-auto">
              <div className="relative">
                <input
                  type="text"
                  placeholder="ค้นหาบทความ..."
                  className="w-full p-3 pl-10 pr-20 text-sm rounded-xl bg-slate-800/50 backdrop-blur-md border border-slate-700/50 dark:text-white placeholder-slate-400 focus:outline-none focus:border-orange-500/50 focus:ring-2 focus:ring-orange-500/20 transition-all duration-300"
                  onChange={handleSearchChange}
                />
                <FiSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-orange-400" />

                {/* Search button */}
                <button className="absolute right-1.5 top-1/2 transform -translate-y-1/2 bg-gradient-to-r from-orange-500 to-red-500 px-3 py-1.5 rounded-lg text-white text-sm font-medium hover:from-orange-600 hover:to-red-600 transition-all duration-300">
                  <FiZap className="w-3 h-3" />
                </button>
              </div>
            </div>
          </div>
        </header>

        {/* Main Content */}
        <section className="mb-12">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center">
              <FiTrendingUp className="w-5 h-5 text-orange-400 mr-2" />
              <h2 className="text-xl font-bold bg-gradient-to-r from-orange-400 to-red-400 bg-clip-text text-transparent">
                Latest From the BSO Blog
              </h2>
            </div>
          </div>

          {/* Blog Posts - Compact layout */}
          <div className="space-y-4">
            {loading && posts?.length === 0 ? (
              <div className="text-center py-8 w-full flex justify-center border-b dark:border-none shadow-sm rounded-md min-h-32 h-full bg-slate-800/50 p-4 text-gray-900 transition-transform transform dark:text-gray-100">
                <div className="w-6 h-6 border-4 border-orange-500 border-t-transparent rounded-full animate-spin"></div>
              </div>
            ) : posts && posts?.length > 0 ? (
              <>
                {posts.map((post, index) => (
                  <div key={`${post.id}-${index}`} className="bg-slate-800/50 backdrop-blur-sm hover:border-slate-600/50 transition-all duration-300">
                    <BlogCard post={post} />
                  </div>
                ))}

                {/* Lazy loading trigger */}
                {hasNextPage && (
                  <div
                    ref={observerRef}
                    className="text-center py-6 w-full flex justify-center"
                  >
                    {loadingMore && (
                      <Loading label="Loading more..." />
                    )}
                  </div>
                )}

                {/* End of content indicator */}
                {!hasNextPage && posts?.length > 5 && (
                  <div className="text-center py-6">
                    <p className="text-slate-400 text-sm">
                      You have already read all stories 🎉
                    </p>
                  </div>
                )}
              </>
            ) : (
              <div className="text-center py-12">
                <div className="mb-3">
                  <FiSearch className="w-12 h-12 text-slate-600 mx-auto" />
                </div>
                <p className="text-base text-slate-400 mb-2">
                  Not found any posts
                </p>
                <p className="text-sm text-slate-500">
                  Try searching for something else or check back later.
                </p>
              </div>
            )}
          </div>
        </section>

        {/* Cookies Consent Modal */}
        <CookiesConsentModal />
      </div>
    </div>
  );
}