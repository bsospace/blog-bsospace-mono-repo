/* eslint-disable @next/next/no-img-element */
"use client";
import { ReactNode, useContext, useEffect, useState, useRef } from "react";
import ThemeSwitcher from "./ThemeSwitcher";
import Image from "next/image";
import logo from "../../public/logo.svg";
import Link from "next/link";
import axios from "axios";
import Script from "next/script";
import { AuthContext } from "../contexts/authContext";
import { ChevronDown, LogOut, Notebook, Settings, SquarePen, User, UserCircle } from "lucide-react";
import { getnerateId } from "@/lib/utils";
import { FiCode, FiCpu } from "react-icons/fi";
import { Button } from "@/components/ui/button";
import NotificationDropdown from "./NotificationDropdown";
import { axiosInstance } from "../utils/api";

export default function Layout({ children }: { children: ReactNode }) {
  const [version, setVersion] = useState<string>("unknown");
  const [isOpen, setIsOpen] = useState(false);
  const [showNavbar, setShowNavbar] = useState(true);
  const { isLoggedIn, user, setIsLoggedIn, setUser } = useContext(AuthContext);
  const desktopDropdownRef = useRef<HTMLDivElement>(null);
  const mobileDropdownRef = useRef<HTMLDivElement>(null);
  const lastScrollY = useRef(0);

  const toggleDropdown = () => {
    setIsOpen(!isOpen);
  };

  const handleLogout = async () => {
    setIsLoggedIn(false);
    setUser(null);
    setIsOpen(false);
    try {
      const response = await axiosInstance.delete("/auth/logout");
      if (response.data.success) {
        // Clear local storage or any other state management
        localStorage.removeItem("warp");
        localStorage.removeItem("logged_in");
        localStorage.removeItem("pid");
        // Redirect to login page
        window.location.href = "/home"
      }
    }
    catch (error) {
      console.error("Error during logout:", error);
    }
  };

  const navigateToLogin = () => {
    window.location.href = "/auth/login";
  };

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node;
      const desktopContains = desktopDropdownRef.current?.contains(target);
      const mobileContains = mobileDropdownRef.current?.contains(target);
      if (!desktopContains && !mobileContains) setIsOpen(false);
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  useEffect(() => {
    const fetchLatestVersion = async () => {
      try {
        const response = await axios.get(
          "https://api.github.com/repos/bsospace/blog-bsospace-mono-repo/releases/latest"
        );
        setVersion(response.data.tag_name || "unknown");
      } catch (error) {
        console.error("Error fetching latest version from GitHub:", error);
        setVersion("unknown");
      }
    };

    fetchLatestVersion();
  }, []);

  // Handle show/hide navbar on scroll (mobile only; always shown on md+)
  useEffect(() => {
    const onScroll = () => {
      const currentY = window.scrollY;
      if (currentY <= 16) {
        setShowNavbar(true);
      } else if (currentY > lastScrollY.current && currentY > 80) {
        setShowNavbar(false);
      } else {
        setShowNavbar(true);
      }
      lastScrollY.current = currentY;
    };

    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <div className="flex flex-col min-h-screen dark:bg-space-dark bg-space-light">
      <Script
        data-name="BMC-Widget"
        data-cfasync="false"
        src="https://cdnjs.buymeacoffee.com/1.0.0/widget.prod.min.js"
        data-id="bsospace"
        data-description="Support me on Buy me a coffee!"
        data-message=""
        data-color="#FF813F"
        data-position="Right"
        data-x_margin="18"
        data-y_margin="18"
      ></Script>

      {/* Header: fixed, rounded, blurred navbar */}
      <div
        className={`fixed top-0 z-50 w-full transition-all duration-300 py-2 px-1 md:py-4 md:px-2 ${showNavbar ? "opacity-100 translate-y-0" : "opacity-0 -translate-y-4"} md:opacity-100 md:translate-y-0`}
      >
        <div className="bg-white/70 dark:bg-gray-900/70 backdrop-blur-sm shadow-sm rounded-full max-w-6xl items-center mx-auto">
          <div className="mx-auto py-1 px-1 md:px-4 md:py-2 flex items-center">
            {/* Logo */}
            <Link href="/" className="flex items-center ms-4 md:ms-0 gap-2 no-underline text-black dark:text-white">
              <Image src={logo} alt="BSO logo" width={32} height={32} />
            </Link>

            {/* Right Controls */}
            <div className="w-full hidden md:flex justify-end items-center gap-4">
              {/* desktop only */}
              <nav className="hidden md:flex md:flex-end space-x-6 text-sm items-center font-medium text-gray-700 dark:text-gray-300">
                <a
                  href="https://github.com/bsospace"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-orange-500 no-underline"
                >
                  GitHub
                </a>
                <a
                  href="https://www.youtube.com/@bsospace"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-orange-500 no-underline"
                >
                  YouTube
                </a>
                <NotificationDropdown />
              </nav>
              {/* Separator */}
              <div className="hidden md:block border-l border-gray-200 dark:border-gray-700 h-6"></div>
              {/* Theme Switcher */}
              <ThemeSwitcher />
              {/* Profile Dropdown */}
              <div className="relative" ref={desktopDropdownRef}>
                <button
                  onClick={toggleDropdown}
                  className="flex items-center gap-2 md:p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors focus:outline-none focus:ring-2 focus:ring-[#fb923c] focus:ring-opacity-50"
                  aria-expanded={isOpen}
                  aria-haspopup="true"
                >
                  {isLoggedIn && user ? (
                    <>
                      <div className="w-8 h-8 rounded-full overflow-hidden border-2 border-white dark:border-gray-700 shadow-sm">
                        <img src={user.avatar} alt="Profile" className="w-full h-full object-cover" />
                      </div>
                      <ChevronDown className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                    </>
                  ) : (
                    <>
                      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 dark:bg-gray-800">
                        <User className="w-5 h-5 text-gray-700 dark:text-gray-300" />
                      </div>
                      <ChevronDown className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                    </>
                  )}
                </button>
                {isOpen && (
                  <div className="absolute right-0 mt-2 w-72 bg-white dark:bg-gray-900 rounded-lg shadow-xl ring-1 ring-black ring-opacity-5 focus:outline-none z-50 origin-top-right transition-all duration-200 ease-out">
                    <div className="p-4">
                      {isLoggedIn && user ? (
                        <>
                          <div className="flex items-center gap-4 pb-4 border-b border-gray-200 dark:border-gray-700">
                            <div className="w-12 h-12 rounded-full overflow-hidden border-2 border-white dark:border-gray-700 shadow">
                              <img src={user.avatar} alt="Profile" className="w-full h-full object-cover" />
                            </div>
                            <div className="flex-1 min-w-0">
                              <p className="font-medium text-lg text-gray-900 dark:text-white truncate leading-tight">
                                {user.username || user.email.split("@")[0]}
                              </p>
                              <p className="text-sm text-gray-500 dark:text-gray-400 truncate leading-tight">{user.email}</p>
                            </div>
                          </div>
                          <div className="mt-4 space-y-1">
                            <Link
                              href={`/w/${getnerateId()}`}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-700 dark:text-gray-300"
                              onClick={() => setIsOpen(false)}
                            >
                              <SquarePen className="w-5 h-5" />
                              <span>Write story</span>
                            </Link>
                            <Link
                              href={`/w`}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-700 dark:text-gray-300"
                              onClick={() => setIsOpen(false)}
                            >
                              <Notebook className="w-5 h-5" />
                              <span>My stories</span>
                            </Link>
                            <Link
                              href={`${"/@"}${user.username}`}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-700 dark:text-gray-300"
                              onClick={(e) => {
                                e.preventDefault();
                                setIsOpen(false);
                              }}
                            >
                              <UserCircle className="w-5 h-5" />
                              <span>Profile</span>
                            </Link>
                            <Link
                              href="/settings"
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-400 dark:text-gray-600 cursor-not-allowed"
                              onClick={(e) => {
                                e.preventDefault();
                                setIsOpen(false);
                              }}
                            >
                              <Settings className="w-5 h-5" />
                              <span>Setting (Comming soon)</span>
                            </Link>
                            <button
                              onClick={handleLogout}
                              className="w-full no-underline flex items-center gap-3 px-3 py-2.5 mt-2 text-sm text-red-600 dark:text-red-400 rounded-md hover:bg-red-50 dark:hover:bg-gray-800 transition-colors"
                            >
                              <LogOut className="w-5 h-5" />
                              <span>Sign out</span>
                            </button>
                          </div>
                        </>
                      ) : (
                        <div className="py-3 text-center text-gray-700 dark:text-gray-300 mb-4">
                          <h3 className="font-semibold text-xl mb-1 bg-gradient-to-r from-orange-400 to-orange-600 text-transparent bg-clip-text">Welcome</h3>
                          <div className="mt-4 space-y-1">
                            <Link
                              href={`/w/${getnerateId()}`}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-700 dark:text-gray-300"
                              onClick={() => setIsOpen(false)}
                            >
                              <SquarePen className="w-5 h-5" />
                              <span>Write story</span>
                            </Link>
                            <Link
                              href={``}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-400 dark:text-gray-600 cursor-not-allowed"
                              onClick={() => setIsOpen(false)}
                            >
                              <Notebook className="w-5 h-5" />
                              <span>My stories</span>
                            </Link>
                            <Link
                              href={``}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-400 dark:text-gray-600 cursor-not-allowed"
                              onClick={(e) => {
                                e.preventDefault();
                                setIsOpen(false);
                              }}
                            >
                              <UserCircle className="w-5 h-5" />
                              <span>Profile</span>
                            </Link>
                            <Link
                              href="/settings"
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-400 dark:text-gray-600 cursor-not-allowed"
                              onClick={(e) => {
                                e.preventDefault();
                                setIsOpen(false);
                              }}
                            >
                              <Settings className="w-5 h-5" />
                              <span>Setting (Comming soon)</span>
                            </Link>
                          </div>
                          <div className="space-y-3 mt-4">
                            <Button variant="default" className="w-full" onClick={navigateToLogin}>
                              Login
                            </Button>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Mobile controls (Theme + Profile) */}
            <div className="md:hidden ml-auto flex items-center gap-2">
              <ThemeSwitcher />
              <div className="relative" ref={mobileDropdownRef}>
                <button
                  onClick={toggleDropdown}
                  className="flex items-center gap-2 p-1.5 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors focus:outline-none focus:ring-2 focus:ring-[#fb923c] focus:ring-opacity-50"
                  aria-expanded={isOpen}
                  aria-haspopup="true"
                >
                  {isLoggedIn && user ? (
                    <>
                      <div className="w-7 h-7 rounded-full overflow-hidden border-2 border-white dark:border-gray-700 shadow-sm">
                        <img src={user.avatar} alt="Profile" className="w-full h-full object-cover" />
                      </div>
                      <ChevronDown className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                    </>
                  ) : (
                    <>
                      <div className="flex items-center justify-center w-7 h-7 rounded-full bg-gray-100 dark:bg-gray-800">
                        <User className="w-4 h-4 text-gray-700 dark:text-gray-300" />
                      </div>
                      <ChevronDown className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                    </>
                  )}
                </button>
                {isOpen && (
                  <div className="absolute right-0 mt-2 w-72 bg-white dark:bg-gray-900 rounded-lg shadow-xl ring-1 ring-black ring-opacity-5 focus:outline-none z-50 origin-top-right transition-all duration-200 ease-out">
                    <div className="p-4">
                      {isLoggedIn && user ? (
                        <>
                          <div className="flex items-center gap-4 pb-4 border-b border-gray-200 dark:border-gray-700">
                            <div className="w-12 h-12 rounded-full overflow-hidden border-2 border-white dark:border-gray-700 shadow">
                              <img src={user.avatar} alt="Profile" className="w-full h-full object-cover" />
                            </div>
                            <div className="flex-1 min-w-0">
                              <p className="font-medium text-lg text-gray-900 dark:text-white truncate leading-tight">
                                {user.username || user.email.split("@")[0]}
                              </p>
                              <p className="text-sm text-gray-500 dark:text-gray-400 truncate leading-tight">{user.email}</p>
                            </div>
                          </div>
                          <div className="mt-4 space-y-1">
                            <Link
                              href={`/w/${getnerateId()}`}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-700 dark:text-gray-300"
                              onClick={() => setIsOpen(false)}
                            >
                              <SquarePen className="w-5 h-5" />
                              <span>Write story</span>
                            </Link>
                            <Link
                              href={`/w`}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-700 dark:text-gray-300"
                              onClick={() => setIsOpen(false)}
                            >
                              <Notebook className="w-5 h-5" />
                              <span>My stories</span>
                            </Link>
                            <Link
                              href={`${"/@"}${user.username}`}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-400 dark:text-gray-600 cursor-not-allowed"
                              onClick={(e) => {
                                e.preventDefault();
                                setIsOpen(false);
                              }}
                            >
                              <UserCircle className="w-5 h-5" />
                              <span>Profile (Coming soon)</span>
                            </Link>
                            <Link
                              href="/settings"
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-400 dark:text-gray-600 cursor-not-allowed"
                              onClick={(e) => {
                                e.preventDefault();
                                setIsOpen(false);
                              }}
                            >
                              <Settings className="w-5 h-5" />
                              <span>Setting (Coming soon)</span>
                            </Link>
                            <button
                              onClick={handleLogout}
                              className="w-full no-underline flex items-center gap-3 px-3 py-2.5 mt-2 text-sm text-red-600 dark:text-red-400 rounded-md hover:bg-red-50 dark:hover:bg-gray-800 transition-colors"
                            >
                              <LogOut className="w-5 h-5" />
                              <span>Sign out</span>
                            </button>
                          </div>
                        </>
                      ) : (
                        <div className="py-3 text-center text-gray-700 dark:text-gray-300 mb-4">
                          <h3 className="font-semibold text-xl mb-1 bg-gradient-to-r from-orange-400 to-orange-600 text-transparent bg-clip-text">Welcome</h3>
                          <div className="mt-4 space-y-1">
                            <Link
                              href={`/w/${getnerateId()}`}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-700 dark:text-gray-300"
                              onClick={() => setIsOpen(false)}
                            >
                              <SquarePen className="w-5 h-5" />
                              <span>Write story</span>
                            </Link>
                            <Link
                              href={`/w`}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-700 dark:text-gray-300"
                              onClick={() => setIsOpen(false)}
                            >
                              <Notebook className="w-5 h-5" />
                              <span>My stories</span>
                            </Link>
                            <Link
                              href={``}
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-400 dark:text-gray-600 cursor-not-allowed"
                              onClick={(e) => {
                                e.preventDefault();
                                setIsOpen(false);
                              }}
                            >
                              <UserCircle className="w-5 h-5" />
                              <span>Profile (Coming soon)</span>
                            </Link>
                            <Link
                              href="/settings"
                              className="flex no-underline items-center gap-3 px-3 py-2.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-400 dark:text-gray-600 cursor-not-allowed"
                              onClick={(e) => {
                                e.preventDefault();
                                setIsOpen(false);
                              }}
                            >
                              <Settings className="w-5 h-5" />
                              <span>Setting (Coming soon)</span>
                            </Link>
                          </div>
                          <div className="space-y-3 mt-4">
                            <Button variant="default" className="w-full" onClick={navigateToLogin}>
                              Login
                            </Button>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <main className="flex-grow w-full mx-auto md:p-6 p-4 pt-20 md:pt-24">

        {/* Background tech elements */}
        {children}
      </main>

      {/* Footer tech elements */}
      <footer className="text-center py-3 border-t border-slate-800">
        <div className="flex justify-center items-center space-x-4 text-slate-400">
          <FiCode className="w-5 h-5 text-orange-400" />
          <span className="text-sm">Be Simple but Outstanding | Version: {version} | &copy; {new Date().getFullYear()} <Link href="https://www.bsospace.com" target="_blank" className="hover:text-orange-400 transition-colors">BSO Space</Link></span>
          <FiCpu className="w-5 h-5 text-red-400" />
        </div>
      </footer>
    </div>
  );
}